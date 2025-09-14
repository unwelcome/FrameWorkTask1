package entities

import "time"

type User struct {
	ID           int64     `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"password_hash"`
	PasswordSalt string    `json:"password_salt"`
	FirstName    string    `json:"first_name"`
	SecondName   string    `json:"second_name"`
	ThirdName    string    `json:"third_name"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	Login      string `json:"login"`
	FirstName  string `json:"first_name"`
	SecondName string `json:"second_name"`
	ThirdName  string `json:"third_name"`
	Role       string `json:"role"`
}

type CreateUserResponse struct {
	ID       int64  `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}
