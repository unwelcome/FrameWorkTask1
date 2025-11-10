package utils

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/unwelcome/FrameWorkTask1/v1/auth/internal/entities"
	"time"
)

// CreateTokens Генерация пары access и refresh токенов
func CreateTokens(userUUID string, secretKey string, accessTokenLifetime, refreshTokenLifetime time.Duration) (*entities.TokenPair, error) {

	// Генерируем access токен
	accessToken, err := generateToken(userUUID, secretKey, entities.AccessTokenType, accessTokenLifetime)
	if err != nil {
		return nil, err
	}

	// Генерируем refresh токен
	refreshToken, err := generateToken(userUUID, secretKey, entities.RefreshTokenType, refreshTokenLifetime)
	if err != nil {
		return nil, err
	}

	// Возвращаем оба токена
	return &entities.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// ParseToken Парсинг jwt токена
func ParseToken(tokenString string, secretKey string) (*entities.TokenClaims, error) {

	// Подтверждаем подлинность токена
	token, err := jwt.ParseWithClaims(tokenString, &entities.TokenClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token expired")
		}
		return nil, fmt.Errorf("failed verify token")
	}

	// Парсим тело токена
	if claims, ok := token.Claims.(*entities.TokenClaims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// generateToken Создание JWT токена
func generateToken(userUUID, secretKey, tokenType string, tokenLifetime time.Duration) (string, error) {
	// Время создания токена
	now := time.Now()

	// Создаем тело токена
	claims := &entities.TokenClaims{
		UserUUID:  userUUID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifetime)),
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
