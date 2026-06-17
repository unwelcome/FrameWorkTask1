package redisDB

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
)

type VerificationRepository interface {
	// TryConsumeVerificationToken атомарно помечает jti токена использованным
	TryConsumeVerificationToken(ctx context.Context, dto entities.ConsumeVerificationTokenDTO) (bool, Error.CodeError)
	// AcquireVerificationEmailCooldown устанавливает cooldown-ключ (SetNX). Возвращает true, если разрешено отправить письмо
	AcquireVerificationEmailCooldown(ctx context.Context, dto entities.AcquireVerificationEmailCooldownDTO) (bool, Error.CodeError)
	// IncrVerificationEmailDailyCount увеличивает суточный счётчик отправок писем верификации и возвращает новое значение
	IncrVerificationEmailDailyCount(ctx context.Context, dto entities.IncrVerificationEmailDailyCountDTO) (int64, Error.CodeError)
}

type verificationRepository struct {
	redis        *redis.Client
	prefix       string
	emailLimiter emailRateLimiter
}

func NewVerificationRepository(rdb *redis.Client, prefix string) VerificationRepository {
	return &verificationRepository{
		redis:        rdb,
		prefix:       prefix,
		emailLimiter: newEmailRateLimiter(rdb, prefix+":verification"),
	}
}

func (r *verificationRepository) TryConsumeVerificationToken(ctx context.Context, dto entities.ConsumeVerificationTokenDTO) (bool, Error.CodeError) {
	ttl := dto.TTL
	if ttl <= 0 {
		ttl = time.Second
	}
	// SetNX атомарно: ставит ключ только если его ещё нет. true → токен claimed впервые.
	claimed, err := r.redis.SetNX(ctx, r.usedKey(dto.TokenID), 1, ttl).Result()
	if err != nil {
		return false, Error.Internal(err)
	}
	return claimed, Error.CodeError{}
}

func (r *verificationRepository) AcquireVerificationEmailCooldown(ctx context.Context, dto entities.AcquireVerificationEmailCooldownDTO) (bool, Error.CodeError) {
	return r.emailLimiter.acquireCooldown(ctx, dto.UserUUID)
}

func (r *verificationRepository) IncrVerificationEmailDailyCount(ctx context.Context, dto entities.IncrVerificationEmailDailyCountDTO) (int64, Error.CodeError) {
	return r.emailLimiter.incrDailyCount(ctx, dto.UserUUID)
}

func (r *verificationRepository) usedKey(tokenID string) string {
	return fmt.Sprintf("%s:verification:%s:used", r.prefix, tokenID)
}
