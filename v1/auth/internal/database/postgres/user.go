package postgresDB

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/unwelcome/FrameWorkTask1/v1/auth/internal/entities"
	"golang.org/x/net/context"
)

type UserRepository interface {
	CreateUser(ctx context.Context, dto *entities.UserCreate) error
	GetUserByEmail(ctx context.Context, email string) (*entities.UserGetByEmail, error)
	GetUser(ctx context.Context, uuid string) (*entities.UserGet, error)
	UpdateUserBio(ctx context.Context, dto *entities.UserUpdateBio) error
	UpdateUserRole(ctx context.Context, dto *entities.UserUpdateRole) error
	DeleteUser(ctx context.Context, uuid string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, dto *entities.UserCreate) error {
	query := `INSERT INTO users (uuid, email, password_hash, first_name, last_name, patronymic) VALUES ($1, $2, $3, $4, $5, $6);`

	_, err := r.db.ExecContext(ctx, query, dto.UserUUID, dto.Email, dto.PasswordHash, dto.FirstName, dto.LastName, dto.Patronymic)
	if err != nil {
		// Проверяем Postgres error code
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "users_uuid_key" {
					return fmt.Errorf("user_uuid already exists")
				} else if pqErr.Constraint == "users_email_key" {
					return fmt.Errorf("email already registered")
				}
			}
		}
		return err
	}
	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*entities.UserGetByEmail, error) {
	query := `SELECT uuid, password_hash FROM users WHERE email = $1;`

	userGetByEmail := &entities.UserGetByEmail{}

	err := r.db.QueryRowContext(ctx, query, email).Scan(&userGetByEmail.UserUUID, &userGetByEmail.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("email not found")
		}
		return nil, err
	}

	return userGetByEmail, nil
}

func (r *userRepository) GetUser(ctx context.Context, uuid string) (*entities.UserGet, error) {
	query := `SELECT role, email, first_name, last_name, patronymic, created_at FROM users WHERE uuid = $1;`

	userGet := &entities.UserGet{UserUUID: uuid}

	err := r.db.QueryRowContext(ctx, query, uuid).Scan(&userGet.Role, &userGet.Email, &userGet.FirstName, &userGet.LastName, &userGet.Patronymic, &userGet.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return userGet, nil
}

func (r *userRepository) UpdateUserBio(ctx context.Context, dto *entities.UserUpdateBio) error {
	query := `UPDATE users SET (first_name, last_name, patronymic) VALUES ($2, $3, $4) WHERE uuid = $1;`

	_, err := r.db.ExecContext(ctx, query, dto.UserUUID, dto.FirstName, dto.LastName, dto.Patronymic)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user not found")
		}
		return err
	}

	return nil
}

func (r *userRepository) UpdateUserRole(ctx context.Context, dto *entities.UserUpdateRole) error {
	query := `UPDATE users SET role = $2 WHERE uuid = $1;`

	_, err := r.db.ExecContext(ctx, query, dto.UserUUID, dto.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user not found")
		}
		return err
	}

	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, uuid string) error {
	query := `DELETE FROM users WHERE uuid = $1;`

	_, err := r.db.ExecContext(ctx, query, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user not found")
		}
		return err
	}

	return nil
}
