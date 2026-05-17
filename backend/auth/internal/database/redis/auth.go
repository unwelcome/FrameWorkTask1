package redisDB

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

type AuthRepository interface {
	SaveRefreshToken(ctx context.Context, userUUID, rawToken string) Error.CodeError
	GetAllRefreshTokens(ctx context.Context, userUUID string) ([]string, Error.CodeError)
	CheckRefreshTokenExists(ctx context.Context, userUUID, rawToken string) Error.CodeError
	RevokeRefreshToken(ctx context.Context, userUUID, tokenHash string) Error.CodeError
	RevokeAllRefreshTokens(ctx context.Context, userUUID string) Error.CodeError
	RefreshToken(ctx context.Context, userUUID, oldRawToken, newRawToken string) Error.CodeError
}

type authRepository struct {
	redis           *redis.Client
	refreshTokenTTL time.Duration
	prefix          string
}

func NewAuthRepository(redis *redis.Client, refreshTokenTTL time.Duration, prefix string) AuthRepository {
	return &authRepository{redis: redis, refreshTokenTTL: refreshTokenTTL, prefix: prefix}
}

func (r *authRepository) SaveRefreshToken(ctx context.Context, userUUID, rawToken string) Error.CodeError {
	hash := r.hashToken(rawToken)

	userTokensKey := r.getUserTokensKey(userUUID)
	tokenKey := r.getRefreshTokenKey(hash)

	pipeline := r.redis.Pipeline()
	pipeline.Set(ctx, tokenKey, 1, r.refreshTokenTTL)
	pipeline.SAdd(ctx, userTokensKey, hash)

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	return Error.CodeError{Code: -1, Err: nil}
}

func (r *authRepository) GetAllRefreshTokens(ctx context.Context, userUUID string) ([]string, Error.CodeError) {
	userTokensKey := r.getUserTokensKey(userUUID)

	hashedTokens, err := r.redis.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return nil, Error.CodeError{Code: 0, Err: err}
	}

	actualRefreshTokens := make([]string, 0)
	expiredTokens := make([]string, 0)

	for _, hash := range hashedTokens {
		tokenKey := r.getRefreshTokenKey(hash)

		count, err := r.redis.Exists(ctx, tokenKey).Result()
		if err != nil {
			continue
		}
		if count == 0 {
			expiredTokens = append(expiredTokens, hash)
			continue
		}

		actualRefreshTokens = append(actualRefreshTokens, hash)
	}

	if len(expiredTokens) > 0 {
		_ = r.redis.SRem(ctx, userTokensKey, expiredTokens).Err()
	}

	return actualRefreshTokens, Error.CodeError{Code: -1, Err: nil}
}

func (r *authRepository) CheckRefreshTokenExists(ctx context.Context, userUUID, rawToken string) Error.CodeError {
	hash := r.hashToken(rawToken)
	tokenKey := r.getRefreshTokenKey(hash)

	exist, err := r.redis.Exists(ctx, tokenKey).Result()
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	if exist == 0 {
		userTokensKey := r.getUserTokensKey(userUUID)
		_ = r.redis.SRem(ctx, userTokensKey, hash).Err()
		return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("refresh token not found")}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

// RevokeRefreshToken принимает хеш токена (не сам токен) и проверяет принадлежность пользователю.
func (r *authRepository) RevokeRefreshToken(ctx context.Context, userUUID, tokenHash string) Error.CodeError {
	userTokensKey := r.getUserTokensKey(userUUID)
	tokenKey := r.getRefreshTokenKey(tokenHash)

	// Проверяем принадлежность токена пользователю
	isMember, err := r.redis.SIsMember(ctx, userTokensKey, tokenHash).Result()
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	if !isMember {
		return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("refresh token not found")}
	}

	// Удаляем ключ токена
	count, err := r.redis.Del(ctx, tokenKey).Result()
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	if count == 0 {
		// Токен уже истек, но ещё числится в сете — чистим
		_ = r.redis.SRem(ctx, userTokensKey, tokenHash).Err()
		return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("refresh token not found")}
	}

	if err := r.redis.SRem(ctx, userTokensKey, tokenHash).Err(); err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

func (r *authRepository) RevokeAllRefreshTokens(ctx context.Context, userUUID string) Error.CodeError {
	userTokensKey := r.getUserTokensKey(userUUID)

	hashedTokens, err := r.redis.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	if len(hashedTokens) == 0 {
		return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("refresh tokens not found")}
	}

	// Удаляем все ключи токенов через pipeline
	pipeline := r.redis.Pipeline()
	for _, hash := range hashedTokens {
		pipeline.Del(ctx, r.getRefreshTokenKey(hash))
	}
	pipeline.Del(ctx, userTokensKey)
	if _, err := pipeline.Exec(ctx); err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}

	return Error.CodeError{Code: -1, Err: nil}
}

// RefreshToken атомарно заменяет старый refresh токен на новый через Watch + TxPipelined.
func (r *authRepository) RefreshToken(ctx context.Context, userUUID, oldRawToken, newRawToken string) Error.CodeError {
	oldHash := r.hashToken(oldRawToken)
	newHash := r.hashToken(newRawToken)

	userTokensKey := r.getUserTokensKey(userUUID)
	oldTokenKey := r.getRefreshTokenKey(oldHash)
	newTokenKey := r.getRefreshTokenKey(newHash)

	err := r.redis.Watch(ctx, func(tx *redis.Tx) error {
		exist, err := tx.Exists(ctx, oldTokenKey).Result()
		if err != nil {
			return err
		}
		if exist == 0 {
			return fmt.Errorf("refresh token not found")
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, newTokenKey, 1, r.refreshTokenTTL)
			pipe.SAdd(ctx, userTokensKey, newHash)
			pipe.Del(ctx, oldTokenKey)
			pipe.SRem(ctx, userTokensKey, oldHash)
			return nil
		})
		return err
	}, oldTokenKey)

	if err != nil {
		if errors.Is(err, fmt.Errorf("refresh token not found")) {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("refresh token not found")}
		}
		return Error.CodeError{Code: 0, Err: err}
	}
	return Error.CodeError{Code: -1, Err: nil}
}

func (r *authRepository) hashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}

func (r *authRepository) getRefreshTokenKey(hash string) string {
	return fmt.Sprintf("%s:token:%s", r.prefix, hash)
}

func (r *authRepository) getUserTokensKey(userUUID string) string {
	return fmt.Sprintf("%s:user:%s:tokens", r.prefix, userUUID)
}
