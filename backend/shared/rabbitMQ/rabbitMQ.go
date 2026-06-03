package rabbitMQ

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const (
	maxRetries = 10
	retryDelay = 3 * time.Second
)

func Connect(connectString string) *amqp.Channel {
	var (
		conn *amqp.Connection
		err  error
	)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		conn, err = amqp.Dial(connectString)
		if err == nil {
			break
		}
		log.Warn().
			Err(err).
			Int("attempt", attempt).
			Int("max", maxRetries).
			Msgf("failed to connect to rabbitMQ, retrying in %s...", retryDelay)
		time.Sleep(retryDelay)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to rabbitMQ after all retries")
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open a channel to rabbitMQ")
	}

	return ch
}
