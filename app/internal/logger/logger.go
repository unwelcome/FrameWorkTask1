package logger

import (
	"github.com/rs/zerolog"
	"os"
)

func MustLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = "02.01.2006 15:04:05.000"
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	return logger
}
