package utils

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenType = "access_token"
)

type TokenClaims struct {
	UserUUID  string `json:"user_uuid"`
	TokenType string `json:"token_type"`
	TokenUUID string `json:"token_uuid"`
	jwt.RegisteredClaims
}

// ParseToken Парсинг и верификация JWT токена по публичному ключу (ES256)
func ParseToken(tokenString string, publicKey *ecdsa.PublicKey) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
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

	if claims, ok := token.Claims.(*TokenClaims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
