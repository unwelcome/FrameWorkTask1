package entities

import (
	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenType        = "access_token"
	RefreshTokenType       = "refresh_token"
	ResetPasswordTokenType = "reset_password_token"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TokenClaims struct {
	TokenUUID string `json:"token_uuid"`
	UserUUID  string `json:"user_uuid"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type ResetPasswordTokenClaims struct {
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}
