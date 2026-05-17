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

	// Path to docker-compose.test.yml relative to this package directory (backend/e2e/).
	// Two levels up reaches the project root (D:\Projects\FrameWorkTask1\).
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
	fmt.Println("=== E2E: gateway ready, running tests ===")

	code := m.Run()

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
