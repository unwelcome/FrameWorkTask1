package redisDB

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
)

type VerificationRepository interface {
	// AddToVerificationTokenBlacklist добавляет jti использованного токена в blacklist.
	// TTL = оставшееся время жизни токена, чтобы Redis самоочищался.
	AddToVerificationTokenBlacklist(ctx context.Context, dto entities.AddToVerificationTokenBlacklistDTO) Error.CodeError
	// IsVerificationTokenBlacklisted проверяет, был ли токен уже использован.
	IsVerificationTokenBlacklisted(ctx context.Context, dto entities.IsVerificationTokenBlacklistedDTO) (bool, Error.CodeError)
}

type verificationRepository struct {
	redis  *redis.Client
	prefix string
}

func NewVerificationRepository(rdb *redis.Client, prefix string) VerificationRepository {
	return &verificationRepository{redis: rdb, prefix: prefix}
}

func (r *verificationRepository) AddToVerificationTokenBlacklist(ctx context.Context, dto entities.AddToVerificationTokenBlacklistDTO) Error.CodeError {
	ttl := dto.TTL
	if ttl <= 0 {
		ttl = time.Second
	}
	if err := r.redis.Set(ctx, r.blacklistKey(dto.TokenID), 1, ttl).Err(); err != nil {
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

func (r *verificationRepository) IsVerificationTokenBlacklisted(ctx context.Context, dto entities.IsVerificationTokenBlacklistedDTO) (bool, Error.CodeError) {
	err := r.redis.Get(ctx, r.blacklistKey(dto.TokenID)).Err()
	if err == nil {
		return true, Error.CodeError{}
	}
	if errors.Is(err, redis.Nil) {
		return false, Error.CodeError{}
	}
	return false, Error.Internal(err)
}

func (r *verificationRepository) blacklistKey(tokenID string) string {
	return fmt.Sprintf("%s:verification:%s:used", r.prefix, tokenID)
}
