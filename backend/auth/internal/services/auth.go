package services

import (
	"context"
	"strings"
	"time"

	"crypto/ecdsa"

	"github.com/google/uuid"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/utils"
	pb "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
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
func (s *AuthService) Health(ctx context.Context, _ *emptypb.Empty) (*pb.HealthResponse, error) {
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
	if err := validate.Email(req.GetEmail()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}
	if err := validate.Password(req.GetPassword()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}
	if err := validate.FirstName(req.GetFirstName()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid first name")
	}
	if err := validate.LastName(req.GetLastName()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid last name")
	}
	if err := validate.Patronymic(req.GetPatronymic()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid patronymic")
	}

	userUUID := uuid.Must(uuid.NewV7()).String()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcryptCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	if err := s.db.User.CreateUser(ctx, &entities.UserCreate{
		UserUUID:     userUUID,
		Email:        req.GetEmail(),
		PasswordHash: string(passwordHash),
		FirstName:    req.GetFirstName(),
		LastName:     req.GetLastName(),
		Patronymic:   req.GetPatronymic(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{UserUuid: userUUID}, nil
}

// Login Авторизация пользователя
func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := validate.Email(req.GetEmail()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}
	if err := validate.Password(req.GetPassword()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}

	user, getErr := s.db.User.GetUserByEmail(ctx, req.GetEmail())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.GetPassword())) != nil {
		return nil, status.Errorf(codes.InvalidArgument, "wrong password")
	}

	tokenPair, err := utils.CreateTokens(user.UserUUID, s.jwtPrivateKey, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	if err := s.cache.Auth.SaveRefreshToken(ctx, user.UserUUID, tokenPair.RefreshToken).GRPCError(); err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		UserUuid:     user.UserUUID,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// GetUser Получение информации о пользователе
func (s *AuthService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	user, getErr := s.db.User.GetUser(ctx, req.GetUserUuid())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

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
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}
	if err := validate.Password(req.GetPassword()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcryptCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	if err := s.db.User.UpdateUserPassword(ctx, req.GetUserUuid(), string(passwordHash)).GRPCError(); err != nil {
		return nil, err
	}

	// Отзываем все токены после смены пароля
	revokeErr := s.cache.Auth.RevokeAllRefreshTokens(ctx, req.GetUserUuid())
	if revokeErr.Code != 0 && revokeErr.Code != int(codes.NotFound) {
		if err := revokeErr.GRPCError(); err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

// UpdateUserBio Обновление ФИО пользователя
func (s *AuthService) UpdateUserBio(ctx context.Context, req *pb.UpdateUserBioRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}
	if err := validate.FirstName(req.GetFirstName()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid first name")
	}
	if err := validate.LastName(req.GetLastName()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid last name")
	}
	if err := validate.Patronymic(req.GetPatronymic()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid patronymic")
	}

	if err := s.db.User.UpdateUserBio(ctx, &entities.UserUpdateBio{
		UserUUID:   req.GetUserUuid(),
		FirstName:  req.GetFirstName(),
		LastName:   req.GetLastName(),
		Patronymic: req.GetPatronymic(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// DeleteUser Удаление пользователя
func (s *AuthService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetTargetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target uuid")
	}

	if req.GetInitiatorUserUuid() != req.GetTargetUserUuid() {
		return nil, status.Errorf(codes.PermissionDenied, "not enough rights")
	}

	// Отзываем все токены пользователя (если они есть)
	revokeErr := s.cache.Auth.RevokeAllRefreshTokens(ctx, req.GetTargetUserUuid())
	if revokeErr.Code != 0 && revokeErr.Code != int(codes.NotFound) {
		if err := revokeErr.GRPCError(); err != nil {
			return nil, err
		}
	}

	if err := s.db.User.DeleteUser(ctx, req.GetTargetUserUuid()).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// GetAllActiveTokens Получение всех активных токенов пользователя
func (s *AuthService) GetAllActiveTokens(ctx context.Context, req *pb.GetAllActiveTokensRequest) (*pb.GetAllActiveTokensResponse, error) {
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	userTokens, getErr := s.cache.Auth.GetAllRefreshTokens(ctx, req.GetUserUuid())
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	tokens := make([]*pb.Token, 0, len(userTokens))
	for _, token := range userTokens {
		tokens = append(tokens, &pb.Token{Token: token})
	}

	return &pb.GetAllActiveTokensResponse{Tokens: tokens}, nil
}

// RefreshToken Обновление токенов
func (s *AuthService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	tokenClaims, err := utils.ParseToken(req.GetRefreshToken(), s.jwtPrivateKey)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if tokenClaims.TokenType != entities.RefreshTokenType {
		return nil, status.Errorf(codes.InvalidArgument, "wrong token type")
	}

	if err = s.cache.Auth.CheckRefreshTokenExists(ctx, tokenClaims.UserUUID, req.GetRefreshToken()).GRPCError(); err != nil {
		return nil, err
	}

	if _, getErr := s.db.User.GetUser(ctx, tokenClaims.UserUUID); getErr.Code != 0 {
		return nil, getErr.GRPCError()
	}

	tokenPair, err := utils.CreateTokens(tokenClaims.UserUUID, s.jwtPrivateKey, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	if err = s.cache.Auth.RefreshToken(ctx, tokenClaims.UserUUID, req.GetRefreshToken(), tokenPair.RefreshToken).GRPCError(); err != nil {
		return nil, err
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RevokeToken Отзыв refresh токена пользователя по его хешу
func (s *AuthService) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}
	if strings.TrimSpace(req.GetTokenHash()) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "token hash missed")
	}

	if err := s.cache.Auth.RevokeRefreshToken(ctx, req.GetUserUuid(), req.GetTokenHash()).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RevokeAllTokens Отзыв всех refresh токенов пользователя
func (s *AuthService) RevokeAllTokens(ctx context.Context, req *pb.RevokeAllTokensRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	if err := s.cache.Auth.RevokeAllRefreshTokens(ctx, req.GetUserUuid()).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
