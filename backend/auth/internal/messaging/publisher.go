package messaging

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/rabbitMQ"
)

type Publisher interface {
	SendVerificationEmail(ctx context.Context, dto entities.VerificationEmailMsg) errors.CodeError
	SendRecoveryEmail(ctx context.Context, dto entities.RecoveryEmailMsg) errors.CodeError
}

type publisher struct {
	ch                     *amqp.Channel
	emailVerificationQueue amqp.Queue
	emailRecoveryQueue     amqp.Queue
}

func NewPublisher(connectString string) Publisher {
	// Подключение к rabbitMQ
	ch := rabbitMQ.Connect(connectString)

	// Создание очереди для email верификации (идемпотентно)
	emailVerificationQueue, err := ch.QueueDeclare(
		"verification.email", // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		amqp.Table{
			amqp.QueueTypeArg: amqp.QueueTypeQuorum,
		},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to declare verification.email queue")
	}

	// Создание очереди для email восстановления пароля (идемпотентно)
	emailRecoveryQueue, err := ch.QueueDeclare(
		"recovery.email", // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		amqp.Table{
			amqp.QueueTypeArg: amqp.QueueTypeQuorum,
		},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to declare recovery.email queue")
	}

	return &publisher{
		ch:                     ch,
		emailVerificationQueue: emailVerificationQueue,
		emailRecoveryQueue:     emailRecoveryQueue,
	}
}

func (p *publisher) SendVerificationEmail(ctx context.Context, dto entities.VerificationEmailMsg) errors.CodeError {
	body, err := json.Marshal(dto)
	if err != nil {
		return errors.Internal(err)
	}

	err = p.ch.PublishWithContext(ctx,
		"",                            // exchange
		p.emailVerificationQueue.Name, // routing key
		false,                         // mandatory
		false,                         // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		})
	if err != nil {
		return errors.Internal(err)
	}

	return errors.CodeError{}
}

func (p *publisher) SendRecoveryEmail(ctx context.Context, dto entities.RecoveryEmailMsg) errors.CodeError {
	body, err := json.Marshal(dto)
	if err != nil {
		return errors.Internal(err)
	}

	err = p.ch.PublishWithContext(ctx,
		"",                        // exchange
		p.emailRecoveryQueue.Name, // routing key
		false,                     // mandatory
		false,                     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		})
	if err != nil {
		return errors.Internal(err)
	}

	return errors.CodeError{}
}
