package utils

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenType  = "access_token"
	RefreshTokenType = "refresh_token"
)

type TokenClaims struct {
	UserUUID  string
	TokenType string
	jwt.RegisteredClaims
}

// ParseToken Парсинг jwt токена
func ParseToken(tokenString string, secretKey string) (*TokenClaims, error) {

	// Подтверждаем подлинность токена
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token expired")
		}
		return nil, fmt.Errorf("failed verify token")
	}

	// Парсим тело токена
	if claims, ok := token.Claims.(*TokenClaims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
