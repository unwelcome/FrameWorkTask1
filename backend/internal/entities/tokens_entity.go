package entities

import (
	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenType  = "access_token"
	RefreshTokenType = "refresh_token"
)

type AuthResponse struct {
	*TokenPair
	UserID int `json:"id"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token" cookie:"access_token"`
	RefreshToken string `json:"refresh_token" cookie:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" cookie:"refresh_token"`
}

type LogoutTokenRequest struct {
	RefreshToken string `json:"refresh_token" cookie:"refresh_token"`
}

type TokenClaims struct {
	UserID int    `json:"user_id"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}
