package redisDB

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	Error "github.com/unwelcome/FrameWorkTask1/v1/auth/pkg/errors"
	"google.golang.org/grpc/codes"
)

type AuthRepository interface {
	SaveRefreshToken(ctx context.Context, userUUID, rawToken string) *Error.CodeError
	GetAllRefreshTokens(ctx context.Context, userUUID string) ([]string, *Error.CodeError)
	CheckRefreshTokenExists(ctx context.Context, userUUID, rawToken string) *Error.CodeError
	RevokeRefreshToken(ctx context.Context, userUUID, rawToken string) *Error.CodeError
	RevokeAllRefreshTokens(ctx context.Context, userUUID string) *Error.CodeError
	RefreshToken(ctx context.Context, userUUID, oldRawToken, newRawToken string) *Error.CodeError
}

type authRepository struct {
	redis           *redis.Client
	refreshTokenTTL time.Duration
}

func NewAuthRepository(redis *redis.Client, refreshTokenTTL time.Duration) AuthRepository {
	return &authRepository{redis: redis, refreshTokenTTL: refreshTokenTTL}
}

func (r *authRepository) SaveRefreshToken(ctx context.Context, userUUID, rawToken string) *Error.CodeError {
	// Хешиурем refresh токен
	hash := r.hashToken(rawToken)

	// Создаем ключи
	userTokensKey := r.getUserTokensKey(userUUID)
	tokenKey := r.getRefreshTokenKey(hash)

	// Создаем транзакцию
	pipeline := r.redis.Pipeline()

	// Сохраняем сам токен с его временем жизни
	pipeline.Set(ctx, tokenKey, 1, r.refreshTokenTTL)
	// Сохраняем токен для пользователя
	pipeline.SAdd(ctx, userTokensKey, hash)
	// Обновляем время жизни сета refresh токенов пользователя
	pipeline.Expire(ctx, userTokensKey, r.refreshTokenTTL)

	// Завершаем транзакцию
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return &Error.CodeError{Code: 0, Err: err}
	}
	return &Error.CodeError{Code: -1, Err: nil}
}

func (r *authRepository) GetAllRefreshTokens(ctx context.Context, userUUID string) ([]string, *Error.CodeError) {
	// Создаем ключ
	userTokensKey := r.getUserTokensKey(userUUID)

	// Получаем все refresh токены пользователя
	hashedTokens, err := r.redis.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return nil, &Error.CodeError{Code: 0, Err: err}
	}

	actualRefreshTokens := make([]string, 0)
	expiredTokens := make([]string, 0)

	// Проверяем токены
	for _, hash := range hashedTokens {
		tokenKey := r.getRefreshTokenKey(hash)

		err = r.redis.Exists(ctx, tokenKey).Err()
		// Если токен истек, то добавляем его в массив истекших токенов
		if err != nil {
			expiredTokens = append(expiredTokens, hash)
			continue
		}

		actualRefreshTokens = append(actualRefreshTokens, hash)
	}

	// Удаляем истекшие токены
	if len(expiredTokens) > 0 {
		_ = r.redis.SRem(ctx, userTokensKey, expiredTokens).Err()
	}

	return actualRefreshTokens, &Error.CodeError{Code: -1, Err: nil}
}

func (r *authRepository) CheckRefreshTokenExists(ctx context.Context, userUUID, rawToken string) *Error.CodeError {
	// Хешиурем refresh токен
	hash := r.hashToken(rawToken)

	// Создаем ключ
	tokenKey := r.getRefreshTokenKey(hash)

	// Проверяем, существует ли токен
	exist, err := r.redis.Exists(ctx, tokenKey).Result()
	if err != nil {
		return &Error.CodeError{Code: 0, Err: err}
	}

	// Проверяем что токен истек
	if exist == 0 {
		// Удаляем токен из токенов пользователя
		userTokensKey := r.getUserTokensKey(userUUID)
		_ = r.redis.SRem(ctx, userTokensKey, hash).Err()

		return &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("refresh token not found")}
	}

	return &Error.CodeError{Code: -1, Err: nil}
}

func (r *authRepository) RevokeRefreshToken(ctx context.Context, userUUID, rawToken string) *Error.CodeError {
	// Хешиурем refresh токен
	hash := r.hashToken(rawToken)

	// Создаем ключи
	userTokensKey := r.getUserTokensKey(userUUID)
	tokenKey := r.getRefreshTokenKey(hash)

	// Удаляем токен
	err1 := r.redis.Del(ctx, tokenKey).Err()
	// Удаляем токен из токенов пользователя
	err2 := r.redis.SRem(ctx, userTokensKey, hash).Err()

	if err1 != nil {
		return &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("refresh token not found")}
	}
	if err2 != nil {
		return &Error.CodeError{Code: 0, Err: err2}
	}

	return &Error.CodeError{Code: -1, Err: nil}
}

func (r *authRepository) RevokeAllRefreshTokens(ctx context.Context, userUUID string) *Error.CodeError {
	// Создаем ключ
	userTokensKey := r.getUserTokensKey(userUUID)

	// Получаем все refresh токены пользователя
	hashedTokens, err := r.redis.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return &Error.CodeError{Code: 0, Err: err}
	}

	// Токенов нет
	if len(hashedTokens) == 0 {
		return &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("refresh tokens not found")}
	}

	// Удаляем каждый refresh токен
	for _, hash := range hashedTokens {
		tokenKey := r.getRefreshTokenKey(hash)
		_ = r.redis.Del(ctx, tokenKey).Err()
	}

	// Удаляем все refresh токены пользователя
	err = r.redis.SRem(ctx, userTokensKey, hashedTokens).Err()
	if err != nil {
		return &Error.CodeError{Code: 0, Err: err}
	}
	return &Error.CodeError{Code: -1, Err: nil}
}

func (r *authRepository) RefreshToken(ctx context.Context, userUUID, oldRawToken, newRawToken string) *Error.CodeError {
	// Хешиурем старый refresh токен
	oldHash := r.hashToken(oldRawToken)
	// Хешиурем новый refresh токен
	newHash := r.hashToken(newRawToken)

	// Создаем ключи
	userTokensKey := r.getUserTokensKey(userUUID)
	oldTokenKey := r.getRefreshTokenKey(oldHash)
	newTokenKey := r.getRefreshTokenKey(newHash)

	// Заменяем токены в транзакции
	pipeline := r.redis.Pipeline()

	// Сохраняем новый refresh токен
	pipeline.Set(ctx, newTokenKey, 1, r.refreshTokenTTL)
	pipeline.SAdd(ctx, userTokensKey, newHash)
	pipeline.Expire(ctx, userTokensKey, r.refreshTokenTTL)

	// Удаляем старый refresh токен
	pipeline.Del(ctx, oldTokenKey)
	pipeline.SRem(ctx, userTokensKey, oldHash)

	// Выполняем транзакцию
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return &Error.CodeError{Code: int(codes.Internal), Err: err}
	}
	return &Error.CodeError{Code: -1, Err: nil}
}

// hashToken Хеширует refresh токен пользователя
func (r *authRepository) hashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}

func (r *authRepository) getRefreshTokenKey(hash string) string {
	return fmt.Sprintf("token:%s", hash)
}

func (r *authRepository) getUserTokensKey(userUUID string) string {
	return fmt.Sprintf("user:%s:tokens", userUUID)
}
