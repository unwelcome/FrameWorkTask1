package postgresDB

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/unwelcome/FrameWorkTask1/v1/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/v1/auth/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
)

type UserRepository interface {
	CreateUser(ctx context.Context, dto *entities.UserCreate) *Error.CodeError
	GetUserByEmail(ctx context.Context, email string) (*entities.UserGetByEmail, *Error.CodeError)
	GetUser(ctx context.Context, uuid string) (*entities.UserGet, *Error.CodeError)
	UpdateUserBio(ctx context.Context, dto *entities.UserUpdateBio) *Error.CodeError
	DeleteUser(ctx context.Context, uuid string) *Error.CodeError
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, dto *entities.UserCreate) *Error.CodeError {
	query := `INSERT INTO users (uuid, email, password_hash, first_name, last_name, patronymic) VALUES ($1, $2, $3, $4, $5, $6);`

	_, err := r.db.ExecContext(ctx, query, dto.UserUUID, dto.Email, dto.PasswordHash, dto.FirstName, dto.LastName, dto.Patronymic)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Почта занята
			if pqErr.Code == "23505" && pqErr.Constraint == "users_email_key" {
				return &Error.CodeError{Code: int(codes.AlreadyExists), Err: fmt.Errorf("email already registered")}
			}
		}
		return &Error.CodeError{Code: 0, Err: err}
	}
	return &Error.CodeError{Code: -1, Err: nil}
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*entities.UserGetByEmail, *Error.CodeError) {
	query := `SELECT uuid, password_hash FROM users WHERE email = $1;`

	userGetByEmail := &entities.UserGetByEmail{}

	err := r.db.QueryRowContext(ctx, query, email).Scan(&userGetByEmail.UserUUID, &userGetByEmail.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("user not found")}
		}
		return nil, &Error.CodeError{Code: 0, Err: err}
	}
	return userGetByEmail, &Error.CodeError{Code: -1, Err: nil}
}

func (r *userRepository) GetUser(ctx context.Context, uuid string) (*entities.UserGet, *Error.CodeError) {
	query := `SELECT email, first_name, last_name, patronymic, created_at FROM users WHERE uuid = $1;`

	userGet := &entities.UserGet{UserUUID: uuid}

	err := r.db.QueryRowContext(ctx, query, uuid).Scan(&userGet.Email, &userGet.FirstName, &userGet.LastName, &userGet.Patronymic, &userGet.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("user not found")}
		}
		return nil, &Error.CodeError{Code: 0, Err: err}
	}
	return userGet, &Error.CodeError{Code: -1, Err: nil}
}

func (r *userRepository) UpdateUserBio(ctx context.Context, dto *entities.UserUpdateBio) *Error.CodeError {
	query := `UPDATE users SET (first_name, last_name, patronymic) = ($2, $3, $4) WHERE uuid = $1;`

	_, err := r.db.ExecContext(ctx, query, dto.UserUUID, dto.FirstName, dto.LastName, dto.Patronymic)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("user not found")}
		}
		return &Error.CodeError{Code: 0, Err: err}
	}
	return &Error.CodeError{Code: -1, Err: nil}
}

func (r *userRepository) DeleteUser(ctx context.Context, uuid string) *Error.CodeError {
	query := `DELETE FROM users WHERE uuid = $1;`

	_, err := r.db.ExecContext(ctx, query, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("user not found")}
		}
		return &Error.CodeError{Code: 0, Err: err}
	}
	return &Error.CodeError{Code: -1, Err: nil}
}
