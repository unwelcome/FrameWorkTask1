package redisDB

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
)

type AuthRepository interface {
	SaveRefreshToken(ctx context.Context, dto entities.SaveRefreshTokenDTO) Error.CodeError
	GetAllRefreshTokens(ctx context.Context, dto entities.GetAllRefreshTokensDTO) ([]string, Error.CodeError)
	CheckRefreshTokenExists(ctx context.Context, dto entities.CheckRefreshTokenExistsDTO) Error.CodeError
	RevokeRefreshToken(ctx context.Context, dto entities.RevokeRefreshTokenDTO) Error.CodeError
	RevokeAllRefreshTokens(ctx context.Context, dto entities.RevokeAllRefreshTokensDTO) Error.CodeError
	RefreshToken(ctx context.Context, dto entities.RefreshTokenDTO) Error.CodeError
}

type authRepository struct {
	redis           *redis.Client
	refreshTokenTTL time.Duration
	prefix          string
}

func NewAuthRepository(redis *redis.Client, refreshTokenTTL time.Duration, prefix string) AuthRepository {
	return &authRepository{redis: redis, refreshTokenTTL: refreshTokenTTL, prefix: prefix}
}

func (r *authRepository) SaveRefreshToken(ctx context.Context, dto entities.SaveRefreshTokenDTO) Error.CodeError {
	hash := r.hashToken(dto.RawToken)

	userTokensKey := r.getUserTokensKey(dto.UserUUID)
	tokenKey := r.getRefreshTokenKey(hash)

	pipeline := r.redis.Pipeline()
	pipeline.Set(ctx, tokenKey, 1, r.refreshTokenTTL)
	pipeline.SAdd(ctx, userTokensKey, hash)

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

func (r *authRepository) GetAllRefreshTokens(ctx context.Context, dto entities.GetAllRefreshTokensDTO) ([]string, Error.CodeError) {
	userTokensKey := r.getUserTokensKey(dto.UserUUID)

	hashedTokens, err := r.redis.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return nil, Error.CodeError{Code: int(codes.Internal), Err: err}
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

	return actualRefreshTokens, Error.CodeError{}
}

func (r *authRepository) CheckRefreshTokenExists(ctx context.Context, dto entities.CheckRefreshTokenExistsDTO) Error.CodeError {
	hash := r.hashToken(dto.RawToken)
	tokenKey := r.getRefreshTokenKey(hash)

	exist, err := r.redis.Exists(ctx, tokenKey).Result()
	if err != nil {
		return Error.Internal(err)
	}

	if exist == 0 {
		userTokensKey := r.getUserTokensKey(dto.UserUUID)
		_ = r.redis.SRem(ctx, userTokensKey, hash).Err()
		return Error.Public(codes.NotFound, "refresh token not found")
	}

	return Error.CodeError{}
}

// RevokeRefreshToken принимает хеш токена (не сам токен) и проверяет принадлежность пользователю.
func (r *authRepository) RevokeRefreshToken(ctx context.Context, dto entities.RevokeRefreshTokenDTO) Error.CodeError {
	userTokensKey := r.getUserTokensKey(dto.UserUUID)
	tokenKey := r.getRefreshTokenKey(dto.TokenHash)

	isMember, err := r.redis.SIsMember(ctx, userTokensKey, dto.TokenHash).Result()
	if err != nil {
		return Error.Internal(err)
	}
	if !isMember {
		return Error.Public(codes.NotFound, "refresh token not found")
	}

	count, err := r.redis.Del(ctx, tokenKey).Result()
	if err != nil {
		return Error.Internal(err)
	}
	if count == 0 {
		// Токен уже истек, но ещё числится в сете — чистим
		_ = r.redis.SRem(ctx, userTokensKey, dto.TokenHash).Err()
		return Error.Public(codes.NotFound, "refresh token not found")
	}

	if err := r.redis.SRem(ctx, userTokensKey, dto.TokenHash).Err(); err != nil {
		return Error.Internal(err)
	}

	return Error.CodeError{}
}

func (r *authRepository) RevokeAllRefreshTokens(ctx context.Context, dto entities.RevokeAllRefreshTokensDTO) Error.CodeError {
	userTokensKey := r.getUserTokensKey(dto.UserUUID)

	hashedTokens, err := r.redis.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return Error.Internal(err)
	}

	if len(hashedTokens) == 0 {
		return Error.Public(codes.NotFound, "refresh tokens not found")
	}

	pipeline := r.redis.Pipeline()
	for _, hash := range hashedTokens {
		pipeline.Del(ctx, r.getRefreshTokenKey(hash))
	}
	pipeline.Del(ctx, userTokensKey)
	if _, err := pipeline.Exec(ctx); err != nil {
		return Error.Internal(err)
	}

	return Error.CodeError{}
}

// RefreshToken атомарно заменяет старый refresh токен на новый через Watch + TxPipelined.
func (r *authRepository) RefreshToken(ctx context.Context, dto entities.RefreshTokenDTO) Error.CodeError {
	oldHash := r.hashToken(dto.OldRawToken)
	newHash := r.hashToken(dto.NewRawToken)

	userTokensKey := r.getUserTokensKey(dto.UserUUID)
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
			return Error.Public(codes.NotFound, "refresh token not found")
		}
		return Error.Internal(err)
	}
	return Error.CodeError{}
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
