package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

const (
	gatewayPort    = "18080"
	gatewayBaseURL = "http://localhost:" + gatewayPort

	composeFile = "../../docker-compose.test.yml"

	startupTimeout = 5 * time.Minute
)

func TestMain(m *testing.M) {
	fmt.Println("=== E2E: starting test environment ===")

	if err := startCompose(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start compose: %v\n", err)
		stopCompose()
		os.Exit(1)
	}

	fmt.Println("=== E2E: waiting for gateway to be ready ===")
	if err := waitForGateway(startupTimeout); err != nil {
		fmt.Fprintf(os.Stderr, "gateway did not become ready: %v\n", err)
		stopCompose()
		os.Exit(1)
	}
	// Небольшая пауза после того, как все health check'и прошли.
	// gRPC-серверы сообщают "healthy" сразу после завершения миграций,
	// но пул соединений к БД ещё не успевает прогреться — первые запросы
	// к более тяжёлым сервисам (company, application) могут падать с 500.
	// 3 секунды достаточно, чтобы connection pool устоялся.
	fmt.Println("=== E2E: warming up services ===")
	time.Sleep(3 * time.Second)
	fmt.Println("=== E2E: running tests ===")

	code := m.Run()

	if code != 0 {
		dumpContainerLogs()
	}

	fmt.Println("=== E2E: tearing down test environment ===")
	stopCompose()
	os.Exit(code)
}

func startCompose() error {
	cmd := exec.Command(
		"docker", "compose",
		"-f", composeFile,
		"-p", "e2e",
		"up", "-d", "--build",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func stopCompose() {
	cmd := exec.Command(
		"docker", "compose",
		"-f", composeFile,
		"-p", "e2e",
		"down", "--volumes", "--remove-orphans",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

// healthResponse mirrors the JSON returned by GET /api/health.
type healthResponse struct {
	Gateway     string        `json:"gateway"`
	Auth        serviceHealth `json:"auth_service"`
	Company     serviceHealth `json:"company_service"`
	Application serviceHealth `json:"application_service"`
}

type serviceHealth struct {
	Service string `json:"service"`
}

// dumpContainerLogs выводит последние 200 строк логов каждого сервисного
// контейнера. Вызывается только при падении тестов, до stopCompose().
func dumpContainerLogs() {
	containers := []string{
		"gateway_test",
		"auth_service_test",
		"company_service_test",
		"application_service_test",
	}

	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("  Container logs (tests failed)")
	fmt.Println("══════════════════════════════════════════════════════════════")

	for _, name := range containers {
		fmt.Printf("\n>>> %s <<<\n", name)
		cmd := exec.Command("docker", "logs", "--tail", "200", name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout // stderr контейнера тоже в stdout чтобы не перемешивать
		if err := cmd.Run(); err != nil {
			fmt.Printf("(could not get logs for %s: %v)\n", name, err)
		}
	}

	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println()
}

func waitForGateway(timeout time.Duration) error {
	healthURL := gatewayBaseURL + "/api/health"
	deadline := time.Now().Add(timeout)
	httpClient := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := httpClient.Get(healthURL)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			time.Sleep(1 * time.Second)
			continue
		}

		var h healthResponse
		if err := json.Unmarshal(body, &h); err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Ждём пока все три сервиса не сообщат что они healthy.
		// Health handler возвращает 200 даже когда сервисы недоступны,
		// поэтому проверяем тело ответа.
		if h.Auth.Service == "healthy" && h.Company.Service == "healthy" && h.Application.Service == "healthy" {
			return nil
		}

		fmt.Printf("=== E2E: waiting for services (auth=%s company=%s application=%s) ===\n",
			h.Auth.Service, h.Company.Service, h.Application.Service)
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("services did not become healthy within %s", timeout)
}
