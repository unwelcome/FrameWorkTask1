package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"crypto/ecdsa"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/utils"
	pb "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/helpers"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/validate"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bcryptCost = 12

type AuthService struct {
	db              *postgresDB.DatabaseRepository
	cache           *redisDB.CacheRepository
	jwtPrivateKey   *ecdsa.PrivateKey
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	pb.UnimplementedAuthServiceServer
}

func NewAuthService(db *postgresDB.DatabaseRepository, cache *redisDB.CacheRepository, jwtPrivateKey *ecdsa.PrivateKey, accessTokenTTL, refreshTokenTTL time.Duration) *AuthService {
	return &AuthService{
		db:              db,
		cache:           cache,
		jwtPrivateKey:   jwtPrivateKey,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// Health Проверка состояния сервиса
func (s *AuthService) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "health").Msg("success")
	return &pb.HealthResponse{
		Service:  "healthy",
		Postgres: helpers.PingStatus(s.db.Ping(ctx)),
		Redis:    helpers.PingStatus(s.cache.Ping(ctx)),
		Minio:    "not implemented",
		Mongo:    "not implemented",
	}, nil
}

// Register Создание нового пользователя
func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Валидации
	if err := validate.Email(req.GetEmail()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "register").Err(fmt.Errorf("invalid email")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}
	if err := validate.Password(req.GetPassword()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "register").Err(fmt.Errorf("invalid password")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}
	if err := validate.FirstName(req.GetFirstName()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "register").Err(fmt.Errorf("invalid first name")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid first name")
	}
	if err := validate.LastName(req.GetLastName()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "register").Err(fmt.Errorf("invalid last name")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid last name")
	}
	if err := validate.Patronymic(req.GetPatronymic()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "register").Err(fmt.Errorf("invalid patronymic")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid patronymic")
	}

	// Создаем uuid для пользователя
	userUUID := uuid.Must(uuid.NewV7()).String()

	// Переводим пароль из строки в срез байт
	bytePassword := []byte(req.GetPassword())

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
	// Валидации
	if err := validate.Email(req.GetEmail()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "login").Err(fmt.Errorf("invalid email")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}
	if err := validate.Password(req.GetPassword()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "login").Err(fmt.Errorf("invalid password")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}

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
	tokenPair, err := utils.CreateTokens(user.UserUUID, s.jwtPrivateKey, s.accessTokenTTL, s.refreshTokenTTL)
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
	// Валидации
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get user").Err(fmt.Errorf("invalid user uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

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

// ChangePassword Обновление пароля пользователя
func (s *AuthService) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*emptypb.Empty, error) {
	// Валидации
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "change password").Err(fmt.Errorf("invalid user uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}
	if err := validate.Password(req.GetPassword()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "change password").Err(fmt.Errorf("invalid password")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}

	// Переводим пароль из строки в срез байт
	bytePassword := []byte(req.GetPassword())

	// Хешируем пароль
	passwordHash, err := bcrypt.GenerateFromPassword(bytePassword, bcryptCost)
	if err != nil {
		log.Error().Str("id", req.GetOperationId()).Str("method", "change password").Err(err).Msg("error")
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Обновление пароля
	updateErr := s.db.User.UpdateUserPassword(ctx, req.GetUserUuid(), string(passwordHash))
	err = Error.HandleError(updateErr, req.GetOperationId(), "change password")
	if err != nil {
		return nil, err
	}

	// Отзываем все токены после смены пароля
	revokeErr := s.cache.Auth.RevokeAllRefreshTokens(ctx, req.GetUserUuid())
	if revokeErr.Code != 0 && revokeErr.Code != int(codes.NotFound) {
		err = Error.HandleError(revokeErr, req.GetOperationId(), "change password")
		if err != nil {
			return nil, err
		}
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "change password").Msg("success")
	return &emptypb.Empty{}, nil
}

// UpdateUserBio Обновление ФИО пользователя
func (s *AuthService) UpdateUserBio(ctx context.Context, req *pb.UpdateUserBioRequest) (*emptypb.Empty, error) {
	// Валидации
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "update user bio").Err(fmt.Errorf("invalid user uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}
	if err := validate.FirstName(req.GetFirstName()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "update user bio").Err(fmt.Errorf("invalid first name")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid first name")
	}
	if err := validate.LastName(req.GetLastName()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "update user bio").Err(fmt.Errorf("invalid last name")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid last name")
	}
	if err := validate.Patronymic(req.GetPatronymic()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "update user bio").Err(fmt.Errorf("invalid patronymic")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid patronymic")
	}

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
	// Валидации
	if err := validate.UUID(req.GetInitiatorUserUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete user").Err(fmt.Errorf("invalid initiator uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetTargetUserUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete user").Err(fmt.Errorf("invalid target uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid target uuid")
	}

	// Проверяем что пользователь обладает правом на удаление
	if !(req.GetInitiatorUserUuid() == req.GetTargetUserUuid()) {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete user").Err(fmt.Errorf("not enough rights")).Msg("error")
		return nil, status.Errorf(codes.PermissionDenied, "not enough rights")
	}

	// Отзываем все токены пользователя (если они есть)
	revokeErr := s.cache.Auth.RevokeAllRefreshTokens(ctx, req.GetTargetUserUuid())
	if revokeErr.Code != 0 && revokeErr.Code != int(codes.NotFound) {
		err := Error.HandleError(revokeErr, req.GetOperationId(), "delete user")
		if err != nil {
			return nil, err
		}
	}

	// Удаляем пользователя
	deleteErr := s.db.User.DeleteUser(ctx, req.GetTargetUserUuid())
	err := Error.HandleError(deleteErr, req.GetOperationId(), "delete user")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "delete user").Msg("success")
	return &emptypb.Empty{}, nil
}

// GetAllActiveTokens Получение всех активных токенов пользователя
func (s *AuthService) GetAllActiveTokens(ctx context.Context, req *pb.GetAllActiveTokensRequest) (*pb.GetAllActiveTokensResponse, error) {
	// Валидации
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get all tokens").Err(fmt.Errorf("invalid user uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	// Получение всех refresh токенов пользователя
	userTokens, getErr := s.cache.Auth.GetAllRefreshTokens(ctx, req.GetUserUuid())
	err := Error.HandleError(getErr, req.GetOperationId(), "get all tokens")
	if err != nil {
		return nil, err
	}

	// Форматирование ответа
	tokens := make([]*pb.Token, 0)
	for _, token := range userTokens {
		tokens = append(tokens, &pb.Token{Token: token})
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get all tokens").Msg("success")
	return &pb.GetAllActiveTokensResponse{Tokens: tokens}, nil
}

// RefreshToken Обновление токенов
func (s *AuthService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	// Парсинг токена
	tokenClaims, err := utils.ParseToken(req.GetRefreshToken(), s.jwtPrivateKey)
	if err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "refresh token").Err(err).Msg("error")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Проверка типа токена
	if tokenClaims.TokenType != entities.RefreshTokenType {
		return nil, status.Errorf(codes.InvalidArgument, "wrong token type")
	}

	// Проверка существования refresh токена
	checkErr := s.cache.Auth.CheckRefreshTokenExists(ctx, tokenClaims.UserUUID, req.GetRefreshToken())
	err = Error.HandleError(checkErr, req.GetOperationId(), "refresh token")
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
	tokenPair, err := utils.CreateTokens(tokenClaims.UserUUID, s.jwtPrivateKey, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "refresh token").Err(err).Msg("error")
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Замена старого refresh токена на новый
	refreshErr := s.cache.Auth.RefreshToken(ctx, tokenClaims.UserUUID, req.GetRefreshToken(), tokenPair.RefreshToken)
	err = Error.HandleError(refreshErr, req.GetOperationId(), "refresh token")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "refresh tokens").Msg("success")
	return &pb.RefreshTokenResponse{AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken}, nil
}

// RevokeToken Отзыв refresh токена пользователя по его хешу
func (s *AuthService) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*emptypb.Empty, error) {
	// Валидации
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "revoke token").Err(fmt.Errorf("invalid user uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}
	if strings.TrimSpace(req.GetTokenHash()) == "" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "revoke token").Err(fmt.Errorf("token hash missed")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "token hash missed")
	}

	// Удаление refresh токена (метод принимает хеш и проверяет принадлежность пользователю)
	revokeErr := s.cache.Auth.RevokeRefreshToken(ctx, req.GetUserUuid(), req.GetTokenHash())
	err := Error.HandleError(revokeErr, req.GetOperationId(), "revoke token")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "revoke token").Msg("success")
	return &emptypb.Empty{}, nil
}

// RevokeAllTokens Отзыв всех refresh токенов пользователя
func (s *AuthService) RevokeAllTokens(ctx context.Context, req *pb.RevokeAllTokensRequest) (*emptypb.Empty, error) {
	// Валидации
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "revoke all tokens").Err(fmt.Errorf("invalid user uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	// Отзыв всех refresh токенов пользователя
	revokeErr := s.cache.Auth.RevokeAllRefreshTokens(ctx, req.GetUserUuid())
	err := Error.HandleError(revokeErr, req.GetOperationId(), "revoke all tokens")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "revoke all tokens").Msg("success")
	return &emptypb.Empty{}, nil
}
