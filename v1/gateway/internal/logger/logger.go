package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/rs/zerolog"
)

func Setup(logPath string, consoleOut bool) *zerolog.Logger {
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
	if !consoleOut {
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

func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		startTime := time.Now()

		err := c.Next()

		logLevel := zerolog.InfoLevel
		if time.Since(startTime) > time.Second*2 {
			logLevel = zerolog.WarnLevel
		}

		log.WithLevel(logLevel).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("duration", int(time.Since(startTime).Milliseconds())).
			Int("status", c.Response().StatusCode()).
			Msg("request")

		return err
	}
}
