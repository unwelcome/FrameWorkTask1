package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/rs/zerolog"
)

// Setup initialises the logger and returns two loggers:
//   - appLogger  — for application-level logs; includes Caller and Timestamp.
//   - httpLogger — for HTTP access logs; includes Timestamp only (Caller is
//     always the middleware line and is therefore useless there).
//
// All other services that only need the app logger can discard the second
// return value with _.
func Setup(logPath string, consoleOut bool) (appLogger *zerolog.Logger, httpLogger *zerolog.Logger) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = "02.01.2006 15:04:05"

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		parts := strings.Split(file, "/")
		if len(parts) > 2 {
			file = strings.Join(parts[len(parts)-2:], "/")
		}
		return fmt.Sprintf("%s:%d", file, line)
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		log.Fatal().Err(err).Msg("failed to create logger directory")
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open logger file")
	}

	var rawWriter io.Writer = logFile
	if consoleOut {
		rawWriter = io.MultiWriter(logFile, os.Stdout)
	}

	// Desired field order for application logs: level → time → id → method → caller → … → message
	// Timestamp() and Caller() are zerolog hooks that normally run last; the
	// orderedWriter corrects their position transparently for every log line.
	appPriority := []string{
		zerolog.LevelFieldName,  // "level"
		"time",                  // "time"
		"id",                    // operation ID
		"method",                // method
		zerolog.CallerFieldName, // "caller"
	}
	appWriter := &orderedWriter{w: rawWriter, priority: appPriority}

	// level → time → caller → custom fields
	app := zerolog.New(appWriter).With().Timestamp().Caller().Logger()

	// No Timestamp() here — the HTTP middleware adds "time" as the first event
	// field explicitly, so the orderedWriter is not needed for http logs.
	http := zerolog.New(rawWriter).With().Logger()

	log.Info().Msg("logger setup complete")
	return &app, &http
}
