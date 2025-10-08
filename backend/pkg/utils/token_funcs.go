package utils

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/unwelcome/iqjtest/internal/entities"
	"time"
)

// Генерация пары access и refresh токенов

func CreateTokens(userID int, secretKey string, accessTokenLifetime, refreshTokenLifetime time.Duration, tokenID *int) (*entities.TokenPair, error) {

	// Генерируем access токен
	accessToken, err := GenerateToken(userID, secretKey, entities.AccessTokenType, accessTokenLifetime, *tokenID)
	*tokenID++
	if err != nil {
		return nil, err
	}

	// Генерируем refresh токен
	refreshToken, err := GenerateToken(userID, secretKey, entities.RefreshTokenType, refreshTokenLifetime, *tokenID)
	*tokenID++
	if err != nil {
		return nil, err
	}

	// Возвращаем оба токена
	return &entities.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// Создание jwt токена

func GenerateToken(userID int, secretKey string, tokenType string, tokenLifetime time.Duration, tokenID int) (string, error) {

	// Время создания токена
	now := time.Now()

	// Создаем тело токена
	claims := &entities.TokenClaims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifetime)),
			ID:        fmt.Sprintf("token-%d", tokenID),
		},
	}

	// Подписываем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("generate token error: %w", err)
	}

	return tokenString, nil
}

// Парсинг jwt токена

func ParseToken(tokenString string, secretKey string) (*entities.TokenClaims, error) {

	// Подтверждаем подлинность токена
	token, err := jwt.ParseWithClaims(tokenString, &entities.TokenClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token expired")
		}
		return nil, fmt.Errorf("can't verify token")
	}

	// Парсим тело токена
	if claims, ok := token.Claims.(*entities.TokenClaims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func GetTokenKey(userID int, tokenType string) string {
	if tokenType == entities.AccessTokenType {
		return fmt.Sprintf("user:%d:access_tokens", userID)
	} else {
		return fmt.Sprintf("user:%d:refresh_tokens", userID)
	}
}
