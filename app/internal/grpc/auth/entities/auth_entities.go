package auth_entities

import "github.com/golang-jwt/jwt/v5"

const (
	AccessTokenType  = "access_token"
	RefreshTokenType = "refresh_token"
)

type User struct {
	ID           int    `json:"id"`
	Login        string `json:"login"`
	PasswordHash string `json:"password"`
	FirstName    string `json:"first_name"`
	SecondName   string `json:"second_name"`
	ThirdName    string `json:"third_name"`
	CreatedAt    string `json:"created_at"`
}

type TokenClaims struct {
	UserID int    `json:"user_id"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Register

type RegisterRequest struct {
	Login      string `json:"login"`
	Password   string `json:"password"`
	FirstName  string `json:"first_name"`
	SecondName string `json:"second_name"`
	ThirdName  string `json:"third_name"`
}

type RegisterResponse struct {
	UserID       int    `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Login

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID       int    `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Get user

type GetUserRequest struct {
	UserID int `json:"user_id"`
}

type GetUserResponse struct {
	UserID     string `json:"user_id"`
	Login      string `json:"login"`
	FirstName  string `json:"first_name"`
	SecondName string `json:"second_name"`
	ThirdName  string `json:"third_name"`
	CreatedAt  string `json:"created_at"`
}

// Refresh tokens

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Update user fio

type UpdateUserRequest struct {
	UserID     int    `json:"user_id"`
	FirstName  string `json:"first_name"`
	SecondName string `json:"second_name"`
	ThirdName  string `json:"third_name"`
}

type UpdateUserResponse struct {
	UserID     int    `json:"user_id"`
	FirstName  string `json:"first_name"`
	SecondName string `json:"second_name"`
	ThirdName  string `json:"third_name"`
}

// Revoke refresh token

type RevokeTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Delete user

type DeleteUserRequest struct {
	UserID int `json:"user_id"`
}
