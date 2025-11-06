package logger

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/config"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func Setup(envConf *config.Config) *zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = "15:04:05 02.01.2006"

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		parts := strings.Split(file, "/")
		if len(parts) > 2 {
			file = strings.Join(parts[len(parts)-2:], "/")
		}
		return fmt.Sprintf("%s:%d", file, line)
	}

	var writer io.Writer
	if envConf.Application.ProductionType == "prod" {
		logPath := envConf.Application.LogPath
		if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
			log.Fatal().Err(err).Msg("failed to create logger directory")
		}

		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open logger file")
		}
		writer = logFile
	} else {
		writer = os.Stdout
	}

	loggerContext := zerolog.New(writer).
		With().
		Caller().
		Timestamp().
		Logger()

	log.Info().Msg("logger setup complete")
	return &loggerContext
}

func RequestLogger(serviceName string) gin.HandlerFunc {
	logger := log.Logger.
		With().
		Str("service", serviceName).
		Logger()
	return func(c *gin.Context) {
		logger.Info().
			Str("method", c.Request.Method).
			Str("url", c.Request.URL.String()).
			Msg("incoming request")

		start := time.Now()
		defer func() {
			if time.Since(start) > time.Second*2 {
				log.Warn().
					Str("method", c.Request.Method).
					Str("url", c.Request.URL.String()).
					Dur("elapsed_ms", time.Since(start)).
					Msg("long response time")
			}
		}()

		c.Next()
	}
}
