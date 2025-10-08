package repositories

import (
	"context"
	"fmt"
	"github.com/unwelcome/iqjtest/pkg/utils"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unwelcome/iqjtest/internal/entities"
)

type AuthRepository interface {
	AddToken(ctx context.Context, userID int, token string, tokenType string, expiresIn time.Duration) error
	CheckExistsToken(ctx context.Context, userID int, token string, tokenType string) error
	ReplaceToken(ctx context.Context, userID int, oldToken, newToken, tokenType string, expiresIn time.Duration) error
	DeleteToken(ctx context.Context, userID int, token string, tokenType string) error
	DeleteAllTokens(ctx context.Context, userID int) error
}

type authRepositoryImpl struct {
	redis *redis.Client
}

func NewAuthRepository(redis *redis.Client) AuthRepository {
	return &authRepositoryImpl{redis: redis}
}

func (r *authRepositoryImpl) AddToken(ctx context.Context, userID int, token string, tokenType string, expiresIn time.Duration) error {
	key := utils.GetTokenKey(userID, tokenType)

	// Добавляем токен в сет
	err := r.redis.SAdd(ctx, key, token).Err()
	if err != nil {
		return fmt.Errorf("failed to add token: %s", err.Error())
	}

	// Устанавливаем TTL на весь сет
	err = r.redis.Expire(ctx, key, expiresIn).Err()
	if err != nil {
		return fmt.Errorf("failed to set TTL to tokens: %s", err.Error())
	}

	return nil
}

func (r *authRepositoryImpl) CheckExistsToken(ctx context.Context, userID int, token string, tokenType string) error {
	key := utils.GetTokenKey(userID, tokenType)

	exists, err := r.redis.SIsMember(ctx, key, token).Result()
	if err != nil {
		return fmt.Errorf("check exists token failed: %s", err.Error())
	}
	if !exists {
		return fmt.Errorf("token does not exist")
	}

	return nil
}

func (r *authRepositoryImpl) ReplaceToken(ctx context.Context, userID int, oldToken, newToken, tokenType string, expiresIn time.Duration) error {
	key := utils.GetTokenKey(userID, tokenType)

	// Проверяем наличие старого токена
	exist, _ := r.redis.SIsMember(ctx, key, oldToken).Result()
	if !exist {
		return fmt.Errorf("token not found")
	}

	// Создаем транзакцию для замены токенов
	_, err := r.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {

		// Удаляем старый токен
		pipe.SRem(ctx, key, oldToken)

		// Добавляем новый токен
		pipe.SAdd(ctx, key, newToken)

		// Обновляем TTL для всего сета
		pipe.Expire(ctx, key, expiresIn)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to replace tokens: %s", err.Error())
	}

	return nil
}

func (r *authRepositoryImpl) DeleteToken(ctx context.Context, userID int, token string, tokenType string) error {
	key := utils.GetTokenKey(userID, tokenType)

	value, err := r.redis.SRem(ctx, key, token).Result()
	if err != nil {
		return fmt.Errorf("failed to delete token: %s", err.Error())
	} else if value == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

func (r *authRepositoryImpl) DeleteAllTokens(ctx context.Context, userID int) error {
	// Удаляем все access токены
	key := utils.GetTokenKey(userID, entities.AccessTokenType)
	err := r.redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete all access tokens: %s", err.Error())
	}

	// Удаляем все refresh токены
	key = utils.GetTokenKey(userID, entities.RefreshTokenType)
	err = r.redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete all refresh tokens: %s", err.Error())
	}

	return nil
}
