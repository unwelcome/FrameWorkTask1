package entities

import (
	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenType  = "access_token"
	RefreshTokenType = "refresh_token"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TokenPairWithUserUUID struct {
	UserUUID     string `db:"uuid"`
	AccessToken  string
	RefreshToken string
}

type TokenClaims struct {
	UserUUID  string
	TokenType string
	jwt.RegisteredClaims
}
