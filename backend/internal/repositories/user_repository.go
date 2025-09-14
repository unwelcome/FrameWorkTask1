package repositories

import (
	"backend/internal/entities"
	"context"
	"database/sql"
	"time"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (login, password_hash, password_salt, first_name, second_name, third_name, email, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, user.Login, user.PasswordHash, user.PasswordSalt, user.FirstName, user.SecondName, user.ThirdName, user.Email, user.Role, time.Now()).Scan(&user.ID)
	if err != nil {
		return err
	}

	return nil
}
