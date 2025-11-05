package entites

type User struct {
	UserID       string `json:"user_id" db:"user_id"`
	Email        string `json:"email" db:"email"`
	PasswordHash string `json:"password_hash" db:"password_hash"`
	FirstName    string `json:"first_name" db:"first_name"`
	LastName     string `json:"last_name" db:"last_name"`
	Patronymic   string `json:"patronymic" db:"patronymic"`
	CreatedAt    string `json:"created_at" db:"created_at"`
}
