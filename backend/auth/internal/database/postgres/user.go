package postgresDB

import (
	"database/sql"
	"errors"
	"time"

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
	UpdateUserBio(ctx context.Context, dto entities.UserUpdateBioDTO) Error.CodeError
	UpdateUser2FA(ctx context.Context, dto entities.UpdateUser2FADTO) Error.CodeError
	DeleteUser(ctx context.Context, dto entities.DeleteUserDTO) Error.CodeError
	RestoreUser(ctx context.Context, dto entities.RestoreUserDTO) Error.CodeError
	AnonymizeExpiredUsers(ctx context.Context, before time.Time) (int64, error)
	SetUserVerified(ctx context.Context, dto entities.SetUserVerifiedDTO) Error.CodeError
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser Создает нового пользователя
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

// GetUserByEmail Возвращает частичные данные пользователя по его email
func (r *userRepository) GetUserByEmail(ctx context.Context, dto entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
	query := `SELECT uuid, password_hash, first_name, is_verified, two_factor_enabled, deleted_at FROM users WHERE email = $1;`

	userGetByEmail := &entities.UserGetByEmail{Email: dto.Email}
	var passwordHash, firstName sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, dto.Email).Scan(
		&userGetByEmail.UserUUID,
		&passwordHash, &firstName,
		&userGetByEmail.IsVerified, &userGetByEmail.Enabled2FA,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, Error.Public(codes.NotFound, "user not found")
		}
		return nil, Error.Internal(err)
	}

	userGetByEmail.PasswordHash = passwordHash.String
	userGetByEmail.FirstName = firstName.String
	if deletedAt.Valid {
		t := deletedAt.Time
		userGetByEmail.DeletedAt = &t
	}
	return userGetByEmail, Error.CodeError{}
}

// GetUser Возвращает данные пользователя по его uuid
func (r *userRepository) GetUser(ctx context.Context, dto entities.GetUserDTO) (*entities.UserGet, Error.CodeError) {
	query := `SELECT email, password_hash, first_name, last_name, patronymic, description, created_at, is_verified, two_factor_enabled, deleted_at FROM users WHERE uuid = $1;`

	userGet := &entities.UserGet{UserUUID: dto.UserUUID}
	var email, passwordHash, firstName, lastName, patronymic, description sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, dto.UserUUID).Scan(
		&email, &passwordHash, &firstName, &lastName, &patronymic, &description,
		&userGet.CreatedAt, &userGet.IsVerified, &userGet.Enabled2FA,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, Error.Public(codes.NotFound, "user not found")
		}
		return nil, Error.Internal(err)
	}

	userGet.Email = email.String
	userGet.PasswordHash = passwordHash.String
	userGet.FirstName = firstName.String
	userGet.LastName = lastName.String
	userGet.Patronymic = patronymic.String
	userGet.Description = description.String
	if deletedAt.Valid {
		t := deletedAt.Time
		userGet.DeletedAt = &t
	}
	return userGet, Error.CodeError{}
}

// UpdateUserPassword Обновляет пароль пользователя
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

// UpdateUserBio Обновляет данные пользователя (ФИО и описание)
func (r *userRepository) UpdateUserBio(ctx context.Context, dto entities.UserUpdateBioDTO) Error.CodeError {
	query := `UPDATE users SET first_name = $2, last_name = $3, patronymic = $4, description = $5 WHERE uuid = $1;`

	result, err := r.db.ExecContext(ctx, query, dto.UserUUID, dto.FirstName, dto.LastName, dto.Patronymic, dto.Description)
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

// UpdateUser2FA Обновляет двухфакторную авторизацию пользователя
func (r *userRepository) UpdateUser2FA(ctx context.Context, dto entities.UpdateUser2FADTO) Error.CodeError {
	query := `UPDATE users SET two_factor_enabled = $2 WHERE uuid = $1;`

	res, err := r.db.ExecContext(ctx, query, dto.UserUUID, dto.TwoFAEnabled)
	if err != nil {
		return Error.Internal(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if rowsAffected == 0 {
		return Error.Public(codes.NotFound, "user not found")
	}
	return Error.CodeError{}
}

// SetUserVerified Помечает пользователя как верифицированного
func (r *userRepository) SetUserVerified(ctx context.Context, dto entities.SetUserVerifiedDTO) Error.CodeError {
	query := `UPDATE users SET is_verified = true WHERE uuid = $1;`

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

// DeleteUser Мягко удаляет пользователя — проставляет deleted_at = NOW()
func (r *userRepository) DeleteUser(ctx context.Context, dto entities.DeleteUserDTO) Error.CodeError {
	query := `UPDATE users SET deleted_at = NOW() WHERE uuid = $1 AND deleted_at IS NULL;`

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

// RestoreUser Восстанавливает мягко удалённый аккаунт.
func (r *userRepository) RestoreUser(ctx context.Context, dto entities.RestoreUserDTO) Error.CodeError {
	query := `UPDATE users SET deleted_at = NULL WHERE uuid = $1 AND deleted_at IS NOT NULL;`

	result, err := r.db.ExecContext(ctx, query, dto.UserUUID)
	if err != nil {
		return Error.Internal(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Error.Internal(err)
	}

	if rowsAffected == 0 {
		return Error.Public(codes.NotFound, "user not found or already active")
	}
	return Error.CodeError{}
}

// AnonymizeExpiredUsers Обнуляет персональные данные пользователей, возвращает количество анонимизированных записей.
func (r *userRepository) AnonymizeExpiredUsers(ctx context.Context, before time.Time) (int64, error) {
	query := `
		UPDATE users SET
			email              = NULL,
			password_hash      = NULL,
			first_name         = NULL,
			last_name          = NULL,
			patronymic         = NULL,
			two_factor_enabled = false
		WHERE deleted_at IS NOT NULL
		  AND deleted_at < $1
		  AND email IS NOT NULL;`

	result, err := r.db.ExecContext(ctx, query, before)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
