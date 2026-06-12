package entities

import "time"

type User struct {
	UserUUID     string `db:"uuid"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	Patronymic   string `db:"patronymic"`
	Enabled2FA   bool   `db:"two_factor_enabled"`
	CreatedAt    string `db:"created_at"`
}

type UserGet struct {
	UserUUID     string     `db:"uuid"`
	Email        string     `db:"email"`          // Пустая строка если аккаунт анонимизирован
	PasswordHash string     `db:"password_hash"`  // Пустая строка если аккаунт анонимизирован
	FirstName    string     `db:"first_name"`     // Пустая строка если аккаунт анонимизирован
	LastName     string     `db:"last_name"`      // Пустая строка если аккаунт анонимизирован
	Patronymic   string     `db:"patronymic"`     // Пустая строка если аккаунт анонимизирован
	Description  string     `db:"description"`
	CreatedAt    string     `db:"created_at"`
	Enabled2FA   bool       `db:"two_factor_enabled"`
	IsVerified   bool       `db:"is_verified"`
	DeletedAt    *time.Time `db:"deleted_at"` // nil если аккаунт активен
}

type UserGetByEmail struct {
	UserUUID     string     `db:"uuid"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	FirstName    string     `db:"first_name"`
	Enabled2FA   bool       `db:"two_factor_enabled"`
	IsVerified   bool       `db:"is_verified"`
	DeletedAt    *time.Time `db:"deleted_at"` // nil если аккаунт активен
}

type UserUpdateBioDTO struct {
	UserUUID    string `db:"uuid"`
	FirstName   string `db:"first_name"`
	LastName    string `db:"last_name"`
	Patronymic  string `db:"patronymic"`
	Description string `db:"description"`
}

type GetUserByEmailDTO struct {
	Email string
}

type GetUserDTO struct {
	UserUUID string
}

type UpdateUserPasswordDTO struct {
	UserUUID     string
	PasswordHash string
}

type DeleteUserDTO struct {
	UserUUID string
}

type RestoreUserDTO struct {
	UserUUID string
}

type SetUserVerifiedDTO struct {
	UserUUID string
}

type UpdateUser2FADTO struct {
	UserUUID     string
	TwoFAEnabled bool
}
