package postgresDB

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
)

type UserRepository interface {
	CreateUser(ctx context.Context, dto entities.User) Error.CodeError
	GetUserByEmail(ctx context.Context, dto entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError)
	GetUser(ctx context.Context, dto entities.GetUserDTO) (*entities.UserGet, Error.CodeError)
	UpdateUserPassword(ctx context.Context, dto entities.UpdateUserPasswordDTO) Error.CodeError
	UpdateUserBio(ctx context.Context, dto entities.UserUpdateBio) Error.CodeError
	DeleteUser(ctx context.Context, dto entities.DeleteUserDTO) Error.CodeError
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, dto entities.User) Error.CodeError {
	query := `INSERT INTO users (uuid, email, password_hash, first_name, last_name, patronymic) VALUES ($1, $2, $3, $4, $5, $6);`

	_, err := r.db.ExecContext(ctx, query, dto.UserUUID, dto.Email, dto.PasswordHash, dto.FirstName, dto.LastName, dto.Patronymic)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" && pqErr.Constraint == "users_email_key" {
				return Error.Public(codes.AlreadyExists, "email already registered")
			}
		}
		return Error.Internal(err)
	}
	return Error.CodeError{}
}

func (r *userRepository) GetUserByEmail(ctx context.Context, dto entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
	query := `SELECT uuid, password_hash FROM users WHERE email = $1;`

	userGetByEmail := &entities.UserGetByEmail{}

	err := r.db.QueryRowContext(ctx, query, dto.Email).Scan(&userGetByEmail.UserUUID, &userGetByEmail.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, Error.Public(codes.NotFound, "user not found")
		}
		return nil, Error.Internal(err)
	}
	return userGetByEmail, Error.CodeError{}
}

func (r *userRepository) GetUser(ctx context.Context, dto entities.GetUserDTO) (*entities.UserGet, Error.CodeError) {
	query := `SELECT email, first_name, last_name, patronymic, created_at FROM users WHERE uuid = $1;`

	userGet := &entities.UserGet{UserUUID: dto.UserUUID}

	err := r.db.QueryRowContext(ctx, query, dto.UserUUID).Scan(&userGet.Email, &userGet.FirstName, &userGet.LastName, &userGet.Patronymic, &userGet.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, Error.Public(codes.NotFound, "user not found")
		}
		return nil, Error.Internal(err)
	}
	return userGet, Error.CodeError{}
}

func (r *userRepository) UpdateUserPassword(ctx context.Context, dto entities.UpdateUserPasswordDTO) Error.CodeError {
	query := `UPDATE users SET password_hash = $2 WHERE uuid = $1;`

	result, err := r.db.ExecContext(ctx, query, dto.UserUUID, dto.PasswordHash)
	if err != nil {
		return Error.Internal(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if rowsAffected == 0 {
		return Error.Public(codes.NotFound, "user not found")
	}
	return Error.CodeError{}
}

func (r *userRepository) UpdateUserBio(ctx context.Context, dto entities.UserUpdateBio) Error.CodeError {
	query := `UPDATE users SET (first_name, last_name, patronymic) = ($2, $3, $4) WHERE uuid = $1;`

	result, err := r.db.ExecContext(ctx, query, dto.UserUUID, dto.FirstName, dto.LastName, dto.Patronymic)
	if err != nil {
		return Error.Internal(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if rowsAffected == 0 {
		return Error.Public(codes.NotFound, "user not found")
	}
	return Error.CodeError{}
}

func (r *userRepository) DeleteUser(ctx context.Context, dto entities.DeleteUserDTO) Error.CodeError {
	query := `DELETE FROM users WHERE uuid = $1;`

	result, err := r.db.ExecContext(ctx, query, dto.UserUUID)
	if err != nil {
		return Error.Internal(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if rowsAffected == 0 {
		return Error.Public(codes.NotFound, "user not found")
	}
	return Error.CodeError{}
}
