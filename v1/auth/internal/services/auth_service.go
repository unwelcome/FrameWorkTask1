package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	pb "github.com/unwelcome/FrameWorkTask1/v1/auth/api"
	postgresDB "github.com/unwelcome/FrameWorkTask1/v1/auth/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/v1/auth/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/v1/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/v1/auth/pkg/errors"
	"github.com/unwelcome/FrameWorkTask1/v1/auth/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

const bcryptCost = 12

type AuthService struct {
	db              *postgresDB.DatabaseRepository
	cache           *redisDB.CacheRepository
	jwtSecretKey    string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	pb.UnimplementedAuthServiceServer
}

func NewAuthService(db *postgresDB.DatabaseRepository, cache *redisDB.CacheRepository, jwtSecretKey string, accessTokenTTL, refreshTokenTTL time.Duration) *AuthService {
	return &AuthService{
		db:              db,
		cache:           cache,
		jwtSecretKey:    jwtSecretKey,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// TODO:
// fix multiple refresh same token
// add check token type

// Health Проверка состояния сервиса
func (s *AuthService) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "health").Msg("request")
	return &pb.HealthResponse{Health: "healthy"}, nil
}

// Register Создание нового пользователя
func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Создаем uuid для пользователя
	userUUID := uuid.New().String()

	// Переводим пароль из строки в срез байт
	bytePassword := []byte(req.GetPassword())

	// Проверяем длину пароля, больше 72 байт библиотека не захеширует
	if len(bytePassword) >= 70 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "register").Err(fmt.Errorf("password too long")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "password too long")
	}

	// Хешируем пароль
	passwordHash, err := bcrypt.GenerateFromPassword(bytePassword, bcryptCost)
	if err != nil {
		log.Error().Str("id", req.GetOperationId()).Str("method", "register").Err(err).Msg("error")
		return nil, status.Errorf(codes.Internal, "internal error")
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

	createErr := s.db.User.CreateUser(ctx, dto)
	err = Error.HandleError(createErr, req.GetOperationId(), "register")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "register").Msg("success")
	return &pb.RegisterResponse{UserUuid: userUUID}, nil
}

// Login Авторизация пользователя
func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Получаем пользователя по email
	user, getErr := s.db.User.GetUserByEmail(ctx, req.GetEmail())
	err := Error.HandleError(getErr, req.GetOperationId(), "login")
	if err != nil {
		return nil, err
	}

	// Проверяем пароль
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.GetPassword())) != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "login").Err(fmt.Errorf("wrong password")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "wrong password")
	}

	// Генерируем токены
	tokenPair, err := utils.CreateTokens(user.UserUUID, s.jwtSecretKey, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		log.Error().Str("id", req.GetOperationId()).Str("method", "login").Err(err).Msg("error")
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Сохраняем refresh токен
	saveErr := s.cache.Auth.SaveRefreshToken(ctx, user.UserUUID, tokenPair.RefreshToken)
	err = Error.HandleError(saveErr, req.GetOperationId(), "login")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "login").Msg("success")
	return &pb.LoginResponse{UserUuid: user.UserUUID, AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken}, nil
}

// GetUser Получение информации о пользователе
func (s *AuthService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// Получаем данные пользователя
	user, getErr := s.db.User.GetUser(ctx, req.GetUserUuid())
	err := Error.HandleError(getErr, req.GetOperationId(), "get user")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get user").Msg("success")
	return &pb.GetUserResponse{
		UserUuid:   user.UserUUID,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Patronymic: user.Patronymic,
		CreatedAt:  user.CreatedAt,
	}, nil
}

// UpdateUserBio Обновление ФИО пользователя
func (s *AuthService) UpdateUserBio(ctx context.Context, req *pb.UpdateUserBioRequest) (*emptypb.Empty, error) {
	// Обновляем ФИО пользователя
	dto := &entities.UserUpdateBio{
		UserUUID:   req.GetUserUuid(),
		FirstName:  req.GetFirstName(),
		LastName:   req.GetLastName(),
		Patronymic: req.GetPatronymic(),
	}

	updateErr := s.db.User.UpdateUserBio(ctx, dto)
	err := Error.HandleError(updateErr, req.GetOperationId(), "update user bio")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "update user bio").Msg("success")
	return &emptypb.Empty{}, nil
}

// DeleteUser Удаление пользователя
func (s *AuthService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	// Проверяем что пользователь обладает правом на удаление
	if !(req.GetInitiatorUserUuid() == req.GetTargetUserUuid()) {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete user").Err(fmt.Errorf("not enough rights")).Msg("error")
		return nil, status.Errorf(codes.PermissionDenied, "not enough rights")
	}

	// Отзываем все токены пользователя
	revokeErr := s.cache.Auth.RevokeAllRefreshTokens(ctx, req.GetTargetUserUuid())
	err := Error.HandleError(revokeErr, req.GetOperationId(), "delete user")
	if err != nil {
		return nil, err
	}

	// Удаляем пользователя
	deleteErr := s.db.User.DeleteUser(ctx, req.GetTargetUserUuid())
	err = Error.HandleError(deleteErr, req.GetOperationId(), "delete user")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "delete user").Msg("success")
	return &emptypb.Empty{}, nil
}

// RefreshToken Обновление токенов
func (s *AuthService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	// Парсинг refresh токена
	tokenClaims, err := utils.ParseToken(req.GetRefreshToken(), s.jwtSecretKey)
	if err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "refresh token").Err(err).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// Проверка существования refresh токена
	cacheErr := s.cache.Auth.CheckRefreshTokenExists(ctx, tokenClaims.UserUUID, req.GetRefreshToken())
	err = Error.HandleError(cacheErr, req.GetOperationId(), "refresh token")
	if err != nil {
		return nil, err
	}

	// Проверка существования пользователя
	_, getErr := s.db.User.GetUser(ctx, tokenClaims.UserUUID)
	err = Error.HandleError(getErr, req.GetOperationId(), "refresh token")
	if err != nil {
		return nil, err
	}

	// Создание новых токенов
	tokenPair, err := utils.CreateTokens(tokenClaims.UserUUID, s.jwtSecretKey, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "refresh token").Err(err).Msg("error")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	// Замена старого refresh токена на новый
	revokeErr := s.cache.Auth.RefreshToken(ctx, tokenClaims.UserUUID, req.GetRefreshToken(), tokenPair.RefreshToken)
	err = Error.HandleError(revokeErr, req.GetOperationId(), "refresh token")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "refresh tokens").Msg("success")
	return &pb.RefreshTokenResponse{AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken}, nil
}

// RevokeToken Отзыв refresh токена
func (s *AuthService) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*emptypb.Empty, error) {
	// Парсим refresh токен
	tokenClaims, err := utils.ParseToken(req.GetRefreshToken(), s.jwtSecretKey)
	if err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "revoke token").Err(err).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// Удаление refresh токена
	revokeErr := s.cache.Auth.RevokeRefreshToken(ctx, tokenClaims.UserUUID, req.GetRefreshToken())
	err = Error.HandleError(revokeErr, req.GetOperationId(), "revoke token")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "revoke token").Msg("success")
	return &emptypb.Empty{}, nil
}

// RevokeAllTokens

// GetAllActiveTokens
