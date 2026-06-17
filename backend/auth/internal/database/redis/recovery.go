package redisDB

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
)

type RecoveryRepository interface {
	// TryConsumeResetToken атомарно помечает reset-password токен использованным.
	TryConsumeResetToken(ctx context.Context, dto entities.ConsumeResetTokenDTO) (bool, Error.CodeError)
	// AcquireRecoveryEmailCooldown устанавливает cooldown-ключ (SetNX). Возвращает true, если разрешено отправить письмо.
	AcquireRecoveryEmailCooldown(ctx context.Context, dto entities.AcquireRecoveryEmailCooldownDTO) (bool, Error.CodeError)
	// IncrRecoveryEmailDailyCount увеличивает суточный счётчик отправок писем восстановления пароля и возвращает новое значение.
	IncrRecoveryEmailDailyCount(ctx context.Context, dto entities.IncrRecoveryEmailDailyCountDTO) (int64, Error.CodeError)
}

type recoveryRepository struct {
	redis        *redis.Client
	prefix       string
	emailLimiter emailRateLimiter
}

func NewRecoveryRepository(rdb *redis.Client, prefix string) RecoveryRepository {
	return &recoveryRepository{
		redis:        rdb,
		prefix:       prefix,
		emailLimiter: newEmailRateLimiter(rdb, prefix+":recovery"),
	}
}

func (r *recoveryRepository) TryConsumeResetToken(ctx context.Context, dto entities.ConsumeResetTokenDTO) (bool, Error.CodeError) {
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

func (r *recoveryRepository) AcquireRecoveryEmailCooldown(ctx context.Context, dto entities.AcquireRecoveryEmailCooldownDTO) (bool, Error.CodeError) {
	return r.emailLimiter.acquireCooldown(ctx, dto.UserUUID)
}

func (r *recoveryRepository) IncrRecoveryEmailDailyCount(ctx context.Context, dto entities.IncrRecoveryEmailDailyCountDTO) (int64, Error.CodeError) {
	return r.emailLimiter.incrDailyCount(ctx, dto.UserUUID)
}

func (r *recoveryRepository) usedKey(tokenID string) string {
	return fmt.Sprintf("%s:reset-password:%s:used", r.prefix, tokenID)
}
