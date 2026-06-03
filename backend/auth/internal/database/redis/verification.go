package redisDB

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

const verificationCodeTTL = 15 * time.Minute

type VerificationRepository interface {
	SaveVerificationCode(ctx context.Context, dto entities.SaveVerificationCodeDTO) Error.CodeError
	GetVerificationCode(ctx context.Context, dto entities.GetVerificationCodeDTO) (string, Error.CodeError)
	DeleteVerificationCode(ctx context.Context, dto entities.DeleteVerificationCodeDTO) Error.CodeError
	IncrVerificationAttempts(ctx context.Context, dto entities.IncrVerificationAttemptsDTO) (int64, Error.CodeError)
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
		if err == redis.Nil {
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

// IncrVerificationAttempts Увеличивает счётчик неверных попыток, возвращает текущее значение
func (r *verificationRepository) IncrVerificationAttempts(ctx context.Context, dto entities.IncrVerificationAttemptsDTO) (int64, Error.CodeError) {
	attemptsKey := r.getAttemptsKey(dto.UserUUID)

	count, err := r.redis.Incr(ctx, attemptsKey).Result()
	if err != nil {
		return 0, Error.Internal(err)
	}
	return count, Error.CodeError{}
}

func (r *verificationRepository) getCodeKey(userUUID string) string {
	return fmt.Sprintf("%s:verify:%s:code", r.prefix, userUUID)
}

func (r *verificationRepository) getAttemptsKey(userUUID string) string {
	return fmt.Sprintf("%s:verify:%s:attempts", r.prefix, userUUID)
}
