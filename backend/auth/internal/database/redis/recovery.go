package redisDB

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
)

type RecoveryRepository interface {
	// AddToResetTokenBlacklist добавляет использованный reset-password токен в blacklist.
	AddToResetTokenBlacklist(ctx context.Context, dto entities.AddToResetTokenBlacklistDTO) Error.CodeError
	// IsResetTokenBlacklisted проверяет, находится ли токен в blacklist.
	IsResetTokenBlacklisted(ctx context.Context, dto entities.IsResetTokenBlacklistedDTO) (bool, Error.CodeError)
}

type recoveryRepository struct {
	redis  *redis.Client
	prefix string
}

func NewRecoveryRepository(redis *redis.Client, prefix string) RecoveryRepository {
	return &recoveryRepository{redis: redis, prefix: prefix}
}

// AddToResetTokenBlacklist Добавляет jti токена в blacklist с TTL равным оставшемуся времени жизни токена
func (r *recoveryRepository) AddToResetTokenBlacklist(ctx context.Context, dto entities.AddToResetTokenBlacklistDTO) Error.CodeError {
	err := r.redis.Set(ctx, r.getResetTokenBlacklistKey(dto.TokenID), 1, dto.TTL).Err()
	if err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// IsResetTokenBlacklisted Проверяет, был ли токен уже использован
func (r *recoveryRepository) IsResetTokenBlacklisted(ctx context.Context, dto entities.IsResetTokenBlacklistedDTO) (bool, Error.CodeError) {
	exists, err := r.redis.Exists(ctx, r.getResetTokenBlacklistKey(dto.TokenID)).Result()
	if err != nil {
		return false, Error.Internal(err)
	}
	return exists > 0, Error.CodeError{}
}

// ─── Вспомогательные функции ──────────────────────────────────────────────────

func (r *recoveryRepository) getResetTokenBlacklistKey(tokenID string) string {
	return fmt.Sprintf("%s:reset-password:%s:used", r.prefix, tokenID)
}
