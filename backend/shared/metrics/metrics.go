package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

// StartServer запускает HTTP-сервер на указанном порту, отдающий /metrics.
// Запускается в фоновой горутине; возвращается немедленно.
func StartServer(port int) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	go func() {
		addr := fmt.Sprintf(":%d", port)
		log.Info().Int("port", port).Msg("metrics server started")
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal().Err(err).Msg("failed to serve metrics server")
		}
	}()
}
