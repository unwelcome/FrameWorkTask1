package entities

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
	UserUUID   string `db:"uuid"`
	Email      string `db:"email"`
	FirstName  string `db:"first_name"`
	LastName   string `db:"last_name"`
	Patronymic string `db:"patronymic"`
	CreatedAt  string `db:"created_at"`
	Enabled2FA bool   `db:"two_factor_enabled"`
	IsVerified bool   `db:"is_verified"`
}

type UserGetByEmail struct {
	UserUUID     string `db:"uuid"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
	FirstName    string `db:"first_name"`
	Enabled2FA   bool   `db:"two_factor_enabled"`
	IsVerified   bool   `db:"is_verified"`
}

type UserUpdateBioDTO struct {
	UserUUID   string `db:"uuid"`
	FirstName  string `db:"first_name"`
	LastName   string `db:"last_name"`
	Patronymic string `db:"patronymic"`
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

type SetUserVerifiedDTO struct {
	UserUUID string
}

type UpdateUser2FADTO struct {
	UserUUID     string
	TwoFAEnabled bool
}
