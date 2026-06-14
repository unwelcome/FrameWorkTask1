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

func (r *recoveryRepository) AcquireRecoveryEmailCooldown(ctx context.Context, dto entities.AcquireRecoveryEmailCooldownDTO) (bool, Error.CodeError) {
	return r.emailLimiter.acquireCooldown(ctx, dto.UserUUID)
}

func (r *recoveryRepository) IncrRecoveryEmailDailyCount(ctx context.Context, dto entities.IncrRecoveryEmailDailyCountDTO) (int64, Error.CodeError) {
	return r.emailLimiter.incrDailyCount(ctx, dto.UserUUID)
}

// ─── Вспомогательные функции ──────────────────────────────────────────────────

func (r *recoveryRepository) getResetTokenBlacklistKey(tokenID string) string {
	return fmt.Sprintf("%s:reset-password:%s:used", r.prefix, tokenID)
}
