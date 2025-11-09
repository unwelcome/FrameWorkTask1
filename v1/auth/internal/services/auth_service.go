package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	pb "github.com/unwelcome/FrameWorkTask1/v1/auth/api"
	postgresDB "github.com/unwelcome/FrameWorkTask1/v1/auth/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/v1/auth/internal/entities"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bcryptCost = 12

type AuthService struct {
	db *postgresDB.DatabaseRepository
	pb.UnimplementedAuthServiceServer
}

func NewAuthService(db *postgresDB.DatabaseRepository) *AuthService {
	return &AuthService{db: db}
}

// Health Проверка состояния сервиса
func (s *AuthService) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "health").Msg("request")
	return &pb.HealthResponse{Health: "healthy"}, nil
}

// Register Создание нового пользователя
func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "register").Msg("request")

	// Создаем uuid для пользователя
	userUUID := uuid.New().String()

	// Переводим пароль из строки в срез байт
	bytePassword := []byte(req.GetPassword())

	// Проверяем длину пароля, больше 72 байт библиотека не захеширует
	if len(bytePassword) >= 70 {
		return nil, fmt.Errorf("create user error: password too long")
	}

	// Хешируем пароль
	passwordHash, err := bcrypt.GenerateFromPassword(bytePassword, bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("create user error: %w", err)
	}

	// Создаем пользователя
	dto := &entities.UserCreate{
		UserUUID:     userUUID,
		Email:        req.GetEmail(),
		PasswordHash: string(passwordHash),
		FirstName:    req.GetFirstName(),
		LastName:     req.GetLastName(),
		Patronymic:   req.GetPatronymic(),
	}

	err = s.db.User.CreateUser(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf("create user error: %w", err)
	}

	// Генерируем токены
	return &pb.RegisterResponse{UserUuid: userUUID, AccessToken: "access_token", RefreshToken: "refresh_token"}, nil
}

// Login Авторизация пользователя
func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "login").Msg("request")

	// Получаем пользователя по email
	user, err := s.db.User.GetUserByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, fmt.Errorf("login error: %w", err)
	}

	// Проверяем пароль
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.GetPassword())) != nil {
		return nil, fmt.Errorf("login error: wrong password")
	}

	// Генерируем токены
	return &pb.LoginResponse{UserUuid: user.UserUUID, AccessToken: "access_token", RefreshToken: "refresh_token"}, nil
}

// GetUser Получение информации о пользователе
func (s *AuthService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "get user").Msg("request")

	// Получаем данные пользователя
	user, err := s.db.User.GetUser(ctx, req.GetUserUuid())
	if err != nil {
		return nil, fmt.Errorf("get user error: %w", err)
	}

	return &pb.GetUserResponse{
		UserUuid:   user.UserUUID,
		Email:      user.Email,
		Role:       user.Role,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Patronymic: user.Patronymic,
		CreatedAt:  user.CreatedAt,
	}, nil
}

// UpdateUserBio Обновление ФИО пользователя
func (s *AuthService) UpdateUserBio(ctx context.Context, req *pb.UpdateUserBioRequest) (*emptypb.Empty, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "update user bio").Msg("request")

	// Обновляем ФИО пользователя
	dto := &entities.UserUpdateBio{
		UserUUID:   req.GetUserUuid(),
		FirstName:  req.GetFirstName(),
		LastName:   req.GetLastName(),
		Patronymic: req.GetPatronymic(),
	}

	err := s.db.User.UpdateUserBio(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf("update user bio error: %w", err)
	}

	return &emptypb.Empty{}, nil
}

// UpdateUserRole Обновление роли пользователя
func (s *AuthService) UpdateUserRole(ctx context.Context, req *pb.UpdateUserRoleRequest) (*emptypb.Empty, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "update user role").Msg("request")

	// Обновляем роль пользователя
	dto := &entities.UserUpdateRole{
		UserUUID: req.GetUserUuid(),
		Role:     req.GetRole(),
	}

	err := s.db.User.UpdateUserRole(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf("update user role error: %w", err)
	}

	return &emptypb.Empty{}, nil
}

// DeleteUser Удаление пользователя
func (s *AuthService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "delete user").Msg("request")

	// Проверяем что пользователь обладает правом на удаление
	if !(req.GetInitiatorUserUuid() == req.GetTargetUserUuid() || req.GetInitiatorUserUuid() == "chief") {
		return nil, fmt.Errorf("delete user error: not enough rights")
	}

	// Отзываем все токены пользователя

	// Удаляем пользователя
	err := s.db.User.DeleteUser(ctx, req.GetTargetUserUuid())
	if err != nil {
		return nil, fmt.Errorf("delete user error: %w", err)
	}

	return &emptypb.Empty{}, nil
}

// RefreshToken Обновление токенов
func (s *AuthService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "refresh tokens").Msg("request")

	// Проверка refresh токена

	// Проверка существования пользователя

	// Удаление старого refresh токена

	// Создание новых токенов

	return &pb.RefreshTokenResponse{AccessToken: "access_token", RefreshToken: "refresh_token"}, nil
}

// RevokeToken Отзыв refresh токена
func (s *AuthService) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*emptypb.Empty, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "revoke token").Msg("request")

	// Удаление refresh токена

	return &emptypb.Empty{}, nil
}
