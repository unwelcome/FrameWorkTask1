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
	verificationCodeTTL = 15 * time.Minute
	resendCooldownTTL   = 60 * time.Second
	resendDailyTTL      = 24 * time.Hour
)

type VerificationRepository interface {
	SaveVerificationCode(ctx context.Context, dto entities.SaveVerificationCodeDTO) Error.CodeError
	GetVerificationCode(ctx context.Context, dto entities.GetVerificationCodeDTO) (string, Error.CodeError)
	DeleteVerificationCode(ctx context.Context, dto entities.DeleteVerificationCodeDTO) Error.CodeError
	IncrVerificationAttempts(ctx context.Context, dto entities.IncrVerificationAttemptsDTO) (int64, Error.CodeError)
	// AcquireResendCooldown устанавливает ключ cooldown (SetNX). Возвращает true, если cooldown
	// успешно установлен (отправка разрешена), false — если cooldown ещё активен.
	AcquireResendCooldown(ctx context.Context, dto entities.CheckResendCooldownDTO) (bool, Error.CodeError)
	// IncrResendDailyCount увеличивает суточный счётчик повторных отправок и возвращает новое значение.
	IncrResendDailyCount(ctx context.Context, dto entities.IncrResendDailyCountDTO) (int64, Error.CodeError)
}

type verificationRepository struct {
	redis  *redis.Client
	prefix string
}

func NewVerificationRepository(rdb *redis.Client, prefix string) VerificationRepository {
	return &verificationRepository{redis: rdb, prefix: prefix}
}

// SaveVerificationCode Сохраняет код верификации и сбрасывает счётчик попыток
func (r *verificationRepository) SaveVerificationCode(ctx context.Context, dto entities.SaveVerificationCodeDTO) Error.CodeError {
	pipe := r.redis.Pipeline()
	pipe.Set(ctx, r.getCodeKey(dto.UserUUID), dto.Code, verificationCodeTTL)
	pipe.Set(ctx, r.getAttemptsKey(dto.UserUUID), 0, verificationCodeTTL) // счётчик живёт столько же, сколько код

	if _, err := pipe.Exec(ctx); err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// GetVerificationCode Возвращает код верификации по UUID пользователя
func (r *verificationRepository) GetVerificationCode(ctx context.Context, dto entities.GetVerificationCodeDTO) (string, Error.CodeError) {
	code, err := r.redis.Get(ctx, r.getCodeKey(dto.UserUUID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", Error.Public(codes.NotFound, "verification code not found or expired")
		}
		return "", Error.Internal(err)
	}
	return code, Error.CodeError{}
}

// DeleteVerificationCode Удаляет код верификации и счётчик попыток
func (r *verificationRepository) DeleteVerificationCode(ctx context.Context, dto entities.DeleteVerificationCodeDTO) Error.CodeError {
	pipe := r.redis.Pipeline()
	pipe.Del(ctx, r.getCodeKey(dto.UserUUID))
	pipe.Del(ctx, r.getAttemptsKey(dto.UserUUID))

	if _, err := pipe.Exec(ctx); err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

// IncrVerificationAttempts Увеличивает счётчик неверных попыток и возвращает текущее значение
func (r *verificationRepository) IncrVerificationAttempts(ctx context.Context, dto entities.IncrVerificationAttemptsDTO) (int64, Error.CodeError) {
	count, err := r.redis.Incr(ctx, r.getAttemptsKey(dto.UserUUID)).Result()
	if err != nil {
		return 0, Error.Internal(err)
	}
	return count, Error.CodeError{}
}

// AcquireResendCooldown Устанавливает cooldown-ключ. Возвращает true, если ключ создан (разрешено отправить),
// false — если ключ уже существует (cooldown ещё активен).
func (r *verificationRepository) AcquireResendCooldown(ctx context.Context, dto entities.CheckResendCooldownDTO) (bool, Error.CodeError) {
	set, err := r.redis.SetNX(ctx, r.getResendCooldownKey(dto.UserUUID), 1, resendCooldownTTL).Result()
	if err != nil {
		return false, Error.Internal(err)
	}
	return set, Error.CodeError{}
}

// IncrResendDailyCount Увеличивает суточный счётчик повторных отправок
func (r *verificationRepository) IncrResendDailyCount(ctx context.Context, dto entities.IncrResendDailyCountDTO) (int64, Error.CodeError) {
	key := r.getResendDailyKey(dto.UserUUID)
	count, err := r.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, Error.Internal(err)
	}
	if count == 1 {
		r.redis.Expire(ctx, key, resendDailyTTL)
	}
	return count, Error.CodeError{}
}

// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ

func (r *verificationRepository) getCodeKey(userUUID string) string {
	return fmt.Sprintf("%s:verify:%s:code", r.prefix, userUUID)
}

func (r *verificationRepository) getAttemptsKey(userUUID string) string {
	return fmt.Sprintf("%s:verify:%s:attempts", r.prefix, userUUID)
}

func (r *verificationRepository) getResendCooldownKey(userUUID string) string {
	return fmt.Sprintf("%s:verify:%s:resend:cooldown", r.prefix, userUUID)
}

func (r *verificationRepository) getResendDailyKey(userUUID string) string {
	return fmt.Sprintf("%s:verify:%s:resend:daily", r.prefix, userUUID)
}
