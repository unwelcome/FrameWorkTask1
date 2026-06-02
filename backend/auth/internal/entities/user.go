package entities

type User struct {
	UserUUID     string `db:"uuid"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	Patronymic   string `db:"patronymic"`
	CreatedAt    string `db:"created_at"`
}

type UserGet struct {
	UserUUID   string `db:"uuid"`
	Email      string `db:"email"`
	FirstName  string `db:"first_name"`
	LastName   string `db:"last_name"`
	Patronymic string `db:"patronymic"`
	CreatedAt  string `db:"created_at"`
}

type UserGetByEmail struct {
	UserUUID     string `db:"uuid"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
}

type UserUpdateBio struct {
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
