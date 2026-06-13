package redisDB

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

const (
	RecoveryCodeTTL           = 15 * time.Minute
	forgotPasswordCooldownTTL = 60 * time.Second
	forgotPasswordDailyTTL    = 24 * time.Hour
)

type RecoveryRepository interface {
	SaveRecoveryCode(ctx context.Context, dto entities.SaveRecoveryCodeDTO) Error.CodeError
	GetRecoveryCode(ctx context.Context, dto entities.GetRecoveryCodeDTO) (string, Error.CodeError)
	DeleteRecoveryCode(ctx context.Context, dto entities.DeleteRecoveryCodeDTO) Error.CodeError
	IncrRecoveryAttempts(ctx context.Context, dto entities.IncrRecoveryAttemptsDTO) (int64, Error.CodeError)
	// AcquireForgotPasswordCooldown устанавливает ключ cooldown (SetNX). Возвращает true, если cooldown
	// успешно установлен (отправка разрешена), false — если cooldown ещё активен.
	AcquireForgotPasswordCooldown(ctx context.Context, dto entities.CheckForgotPasswordCooldownDTO) (bool, Error.CodeError)
	// IncrForgotPasswordDailyCount увеличивает суточный счётчик запросов восстановления и возвращает новое значение.
	IncrForgotPasswordDailyCount(ctx context.Context, dto entities.IncrForgotPasswordDailyCountDTO) (int64, Error.CodeError)
}

type recoveryRepository struct {
	redis  *redis.Client
	prefix string
}

func NewRecoveryRepository(redis *redis.Client, prefix string) RecoveryRepository {
	return &recoveryRepository{redis: redis, prefix: prefix}
}

// SaveRecoveryCode Сохраняет код восстановления и сбрасывает счётчик попыток
func (r *recoveryRepository) SaveRecoveryCode(ctx context.Context, dto entities.SaveRecoveryCodeDTO) Error.CodeError {
	pipe := r.redis.Pipeline()
	pipe.Set(ctx, r.getRecoveryCodeKey(dto.UserUUID), dto.Code, RecoveryCodeTTL)
	pipe.Set(ctx, r.getRecoveryAttemptsKey(dto.UserUUID), 0, RecoveryCodeTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// GetRecoveryCode Возвращает код восстановления по UUID пользователя
func (r *recoveryRepository) GetRecoveryCode(ctx context.Context, dto entities.GetRecoveryCodeDTO) (string, Error.CodeError) {
	code, err := r.redis.Get(ctx, r.getRecoveryCodeKey(dto.UserUUID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", Error.Public(codes.NotFound, "recovery code not found or expired")
		}
		return "", Error.Internal(err)
	}
	return code, Error.CodeError{}
}

// DeleteRecoveryCode Удаляет код восстановления и счётчик попыток
func (r *recoveryRepository) DeleteRecoveryCode(ctx context.Context, dto entities.DeleteRecoveryCodeDTO) Error.CodeError {
	pipe := r.redis.Pipeline()
	pipe.Del(ctx, r.getRecoveryCodeKey(dto.UserUUID))
	pipe.Del(ctx, r.getRecoveryAttemptsKey(dto.UserUUID))

	if _, err := pipe.Exec(ctx); err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// IncrRecoveryAttempts Увеличивает счётчик неверных попыток восстановления и возвращает текущее значение
func (r *recoveryRepository) IncrRecoveryAttempts(ctx context.Context, dto entities.IncrRecoveryAttemptsDTO) (int64, Error.CodeError) {
	count, err := r.redis.Incr(ctx, r.getRecoveryAttemptsKey(dto.UserUUID)).Result()
	if err != nil {
		return 0, Error.Internal(err)
	}
	return count, Error.CodeError{}
}

// AcquireForgotPasswordCooldown Устанавливает cooldown-ключ. Возвращает true, если ключ создан (разрешено отправить),
// false — если ключ уже существует (cooldown ещё активен).
func (r *recoveryRepository) AcquireForgotPasswordCooldown(ctx context.Context, dto entities.CheckForgotPasswordCooldownDTO) (bool, Error.CodeError) {
	set, err := r.redis.SetNX(ctx, r.getForgotPasswordCooldownKey(dto.UserUUID), 1, forgotPasswordCooldownTTL).Result()
	if err != nil {
		return false, Error.Internal(err)
	}
	return set, Error.CodeError{}
}

// IncrForgotPasswordDailyCount Увеличивает суточный счётчик запросов восстановления пароля
func (r *recoveryRepository) IncrForgotPasswordDailyCount(ctx context.Context, dto entities.IncrForgotPasswordDailyCountDTO) (int64, Error.CodeError) {
	key := r.getForgotPasswordDailyKey(dto.UserUUID)
	count, err := r.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, Error.Internal(err)
	}
	if count == 1 {
		r.redis.Expire(ctx, key, forgotPasswordDailyTTL)
	}
	return count, Error.CodeError{}
}

// ВСПОМОГАТЕЛЬНЫЕ ФУНЦИИ

func (r *recoveryRepository) getRecoveryCodeKey(userUUID string) string {
	return fmt.Sprintf("%s:recovery:%s:code", r.prefix, userUUID)
}

func (r *recoveryRepository) getRecoveryAttemptsKey(userUUID string) string {
	return fmt.Sprintf("%s:recovery:%s:attempts", r.prefix, userUUID)
}

func (r *recoveryRepository) getForgotPasswordCooldownKey(userUUID string) string {
	return fmt.Sprintf("%s:recovery:%s:forgot:cooldown", r.prefix, userUUID)
}

func (r *recoveryRepository) getForgotPasswordDailyKey(userUUID string) string {
	return fmt.Sprintf("%s:recovery:%s:forgot:daily", r.prefix, userUUID)
}
