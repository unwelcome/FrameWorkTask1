package services

import (
	"context"
	"fmt"

	"github.com/unwelcome/iqjtest/internal/entities"
	"github.com/unwelcome/iqjtest/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, userCreate *entities.UserCreateRequest) (int, error)
	LoginUser(ctx context.Context, userLogin *entities.UserLoginRequest) (int, error)
	GetUserByID(ctx context.Context, userID int) (*entities.UserGet, error)
	GetAllUsers(ctx context.Context) ([]*entities.UserGet, error)
	UpdateUserPassword(ctx context.Context, userID int, userUpdatePasswordRequest *entities.UserUpdatePasswordRequest) error
	DeleteUser(ctx context.Context, userID int) error
}

type userServiceImpl struct {
	userRepository repositories.UserRepository
	bcryptCost     int
}

func NewUserService(userRepository repositories.UserRepository, bcryptCost int) UserService {
	return &userServiceImpl{userRepository: userRepository, bcryptCost: bcryptCost}
}

func (s *userServiceImpl) CreateUser(ctx context.Context, userCreate *entities.UserCreateRequest) (int, error) {

	// Переводим пароль из строки в срез байт
	bytePassword := []byte(userCreate.Password)

	// Проверяем длину пароля, больше 72 байт библиотека не захеширует
	if len(bytePassword) >= 70 {
		return 0, fmt.Errorf("create user error: password too long")
	}

	// Хешируем пароль
	passwordHash, err := bcrypt.GenerateFromPassword(bytePassword, s.bcryptCost)
	if err != nil {
		return 0, fmt.Errorf("create user error: %w", err)
	}

	// Создаем пользователя
	user := &entities.User{Login: userCreate.Login, PasswordHash: string(passwordHash)}

	// Добавляем пользователя в бд
	err = s.userRepository.CreateUser(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("create user error: %w", err)
	}

	return user.ID, nil
}

func (s *userServiceImpl) LoginUser(ctx context.Context, userLogin *entities.UserLoginRequest) (int, error) {

	// Получаем пользователя с данным логином
	userWithLogin, err := s.userRepository.GetUserByLogin(ctx, userLogin.Login)
	if err != nil {
		return 0, fmt.Errorf("login user error: user not found")
	}

	// Проверяем пароль
	if bcrypt.CompareHashAndPassword([]byte(userWithLogin.PasswordHash), []byte(userLogin.Password)) != nil {
		return 0, fmt.Errorf("login user error: invalid password")
	}

	return userWithLogin.ID, nil
}

func (s *userServiceImpl) GetUserByID(ctx context.Context, userID int) (*entities.UserGet, error) {

	// Получаем пользователя по ID
	user, err := s.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id error: %w", err)
	}

	return user, nil
}

func (s *userServiceImpl) GetAllUsers(ctx context.Context) ([]*entities.UserGet, error) {

	// Получаем всех пользователей
	users, err := s.userRepository.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all users error: %w", err)
	}

	return users, nil
}

func (s *userServiceImpl) UpdateUserPassword(ctx context.Context, userID int, userUpdatePasswordRequest *entities.UserUpdatePasswordRequest) error {

	// Хешируем новый пароль
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(userUpdatePasswordRequest.Password), s.bcryptCost)
	if err != nil {
		return fmt.Errorf("update user password error: %w", err)
	}

	// Обновляем пароль
	err = s.userRepository.UpdateUserPassword(ctx, userID, string(passwordHash))
	if err != nil {
		return fmt.Errorf("update user password error: %w", err)
	}

	return nil
}

func (s *userServiceImpl) DeleteUser(ctx context.Context, userID int) error {

	// Удаляем пользователя
	err := s.userRepository.DeleteUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("delete user error: %w", err)
	}

	return nil
}
