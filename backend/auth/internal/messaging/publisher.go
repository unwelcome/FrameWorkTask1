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
	Send2FAEmail(ctx context.Context, dto entities.TwoFAEmailMsg) errors.CodeError
	SendPasswordChangedEmail(ctx context.Context, dto entities.PasswordChangedEmailMsg) errors.CodeError
}

type publisher struct {
	ch                          *amqp.Channel
	emailVerificationQueue      amqp.Queue
	emailRecoveryQueue          amqp.Queue
	email2FAQueue               amqp.Queue
	emailPasswordChangedQueue   amqp.Queue
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

	// Создание очереди для 2FA верификации (идемпотентно)
	email2FAQueue, err := ch.QueueDeclare(
		"2fa.email",
		true,
		false,
		false,
		false,
		amqp.Table{
			amqp.QueueTypeArg: amqp.QueueTypeQuorum,
		},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to declare 2fa.email queue")
	}

	// Создание очереди для уведомления о смене пароля (идемпотентно)
	emailPasswordChangedQueue, err := ch.QueueDeclare(
		"password-changed.email",
		true,
		false,
		false,
		false,
		amqp.Table{
			amqp.QueueTypeArg: amqp.QueueTypeQuorum,
		},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to declare password-changed.email queue")
	}

	return &publisher{
		ch:                        ch,
		emailVerificationQueue:    emailVerificationQueue,
		emailRecoveryQueue:        emailRecoveryQueue,
		email2FAQueue:             email2FAQueue,
		emailPasswordChangedQueue: emailPasswordChangedQueue,
	}
}

// SendVerificationEmail Отправляет в очередь verification.email письмо для верификации аккаунта пользователя
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

// SendRecoveryEmail Отправляет в очередь recovery.email письмо для восстановления пароля пользователя
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

// SendPasswordChangedEmail Отправляет в очередь password-changed.email уведомление о смене пароля
func (p *publisher) SendPasswordChangedEmail(ctx context.Context, dto entities.PasswordChangedEmailMsg) errors.CodeError {
	body, err := json.Marshal(dto)
	if err != nil {
		return errors.Internal(err)
	}

	err = p.ch.PublishWithContext(ctx,
		"",
		p.emailPasswordChangedQueue.Name,
		false,
		false,
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

// Send2FAEmail Отправляет в очередь 2fa.email письмо для 2FA авторизации пользователя
func (p *publisher) Send2FAEmail(ctx context.Context, dto entities.TwoFAEmailMsg) errors.CodeError {
	body, err := json.Marshal(dto)
	if err != nil {
		return errors.Internal(err)
	}

	err = p.ch.PublishWithContext(ctx,
		"",
		p.email2FAQueue.Name,
		false,
		false,
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
