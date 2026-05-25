package utils

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
)

// CreateTokens Генерация пары access и refresh токенов
func CreateTokens(userUUID string, privateKey *ecdsa.PrivateKey, accessTokenLifetime, refreshTokenLifetime time.Duration) (*entities.TokenPair, error) {

	accessToken, err := generateToken(userUUID, privateKey, entities.AccessTokenType, accessTokenLifetime)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateToken(userUUID, privateKey, entities.RefreshTokenType, refreshTokenLifetime)
	if err != nil {
		return nil, err
	}

	return &entities.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// ParseToken Парсинг jwt токена (верификация публичным ключом, извлечённым из приватного)
func ParseToken(tokenString string, privateKey *ecdsa.PrivateKey) (*entities.TokenClaims, error) {
	return parseTokenWithPublicKey(tokenString, &privateKey.PublicKey)
}

func parseTokenWithPublicKey(tokenString string, publicKey *ecdsa.PublicKey) (*entities.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &entities.TokenClaims{}, func(token *jwt.Token) (any, error) {
		// Проверяем алгоритм — защита от подмены alg=none или RS256
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token expired")
		}
		return nil, fmt.Errorf("failed verify token")
	}

	if claims, ok := token.Claims.(*entities.TokenClaims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// generateToken Создание JWT токена с ES256
func generateToken(userUUID string, privateKey *ecdsa.PrivateKey, tokenType string, tokenLifetime time.Duration) (string, error) {
	claims := &entities.TokenClaims{
		TokenUUID: uuid.Must(uuid.NewV7()).String(),
		UserUUID:  userUUID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenLifetime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("generate token error: %w", err)
	}

	return tokenString, nil
}
