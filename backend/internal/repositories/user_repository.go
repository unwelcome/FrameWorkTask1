package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/unwelcome/iqjtest/internal/entities"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByID(ctx context.Context, id int) (*entities.UserGet, error)
	GetUserByLogin(ctx context.Context, login string) (*entities.User, error)
	GetAllUsers(ctx context.Context) ([]*entities.UserGet, error)
	UpdateUserPassword(ctx context.Context, id int, passwordHash string) error
	DeleteUser(ctx context.Context, id int) error
}

type userRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) CreateUser(ctx context.Context, user *entities.User) error {
	query := `INSERT INTO users(login, password_hash) VALUES ($1, $2) RETURNING id`

	err := r.db.QueryRowContext(ctx, query, user.Login, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepositoryImpl) GetUserByID(ctx context.Context, id int) (*entities.UserGet, error) {
	query := `SELECT login, created_at FROM users WHERE id = $1`

	// Получаем пользователя по ID
	row := r.db.QueryRowContext(ctx, query, id)

	// Меппинг запроса в структуру
	user := &entities.UserGet{ID: id}
	err := row.Scan(&user.Login, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepositoryImpl) GetUserByLogin(ctx context.Context, login string) (*entities.User, error) {
	query := `SELECT id, password_hash FROM users WHERE login = $1`

	// Получаем пользователя по login
	row := r.db.QueryRowContext(ctx, query, login)

	// Меппинг запроса в структуру
	user := &entities.User{Login: login}
	err := row.Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepositoryImpl) GetAllUsers(ctx context.Context) ([]*entities.UserGet, error) {
	query := `SELECT id, login, created_at FROM users`

	// Получаем всех пользователей
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entities.UserGet

	// Меппим каждого пользователя в структуру
	for rows.Next() {
		user := &entities.UserGet{}
		err = rows.Scan(&user.ID, &user.Login, &user.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Добавляем в массив пользователей
		users = append(users, user)
	}

	return users, nil
}

func (r *userRepositoryImpl) UpdateUserPassword(ctx context.Context, id int, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`

	// Обновляем пароль пользователя
	_, err := r.db.ExecContext(ctx, query, passwordHash, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepositoryImpl) DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	// Удаляем пользователя
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Проверяем что пользователь был удалён
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	} else if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
