package utils

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
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

// parseTokenWithPublicKey Парсинг токена с помощью открытого ключа
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

// CreateResetPasswordToken Генерация JWT токена для сброса пароля
func CreateResetPasswordToken(email string, privateKey *ecdsa.PrivateKey, ttl time.Duration) (string, error) {
	claims := &entities.ResetPasswordTokenClaims{
		Email:     email,
		TokenType: entities.ResetPasswordTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.Must(uuid.NewV7()).String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("generate reset password token error: %w", err)
	}
	return tokenString, nil
}

// ParseResetPasswordToken Парсинг JWT токена для сброса пароля
func ParseResetPasswordToken(tokenString string, privateKey *ecdsa.PrivateKey) (*entities.ResetPasswordTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &entities.ResetPasswordTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &privateKey.PublicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token expired")
		}
		return nil, fmt.Errorf("failed verify token")
	}

	if claims, ok := token.Claims.(*entities.ResetPasswordTokenClaims); ok {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

// CreateVerificationToken Генерация JWT токена для верификации аккаунта
func CreateVerificationToken(email string, privateKey *ecdsa.PrivateKey, ttl time.Duration) (string, error) {
	claims := &entities.VerificationTokenClaims{
		Email:     email,
		TokenType: entities.VerificationTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.Must(uuid.NewV7()).String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("generate verification token error: %w", err)
	}
	return tokenString, nil
}

// ParseVerificationToken Парсинг JWT токена для верификации аккаунта
func ParseVerificationToken(tokenString string, privateKey *ecdsa.PrivateKey) (*entities.VerificationTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &entities.VerificationTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &privateKey.PublicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token expired")
		}
		return nil, fmt.Errorf("failed verify token")
	}
	if claims, ok := token.Claims.(*entities.VerificationTokenClaims); ok {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

// HashToken Хеширует refresh токен
func HashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}
