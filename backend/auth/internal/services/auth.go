package services

import (
	"context"
	"crypto/ecdsa"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/messaging"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/password"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/utils"
	pb "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/format"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/helpers"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/interceptors"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/validate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	maxVerificationAttempts = 5
	maxRecoveryAttempts     = 5
	max2FAAttempts          = 5

	// accountDeletionRetention — срок хранения данных мягко удалённого аккаунта
	accountDeletionRetention = 30 * 24 * time.Hour

	// dummyUUID — nil UUID, используется как заглушка в Redis-запросах при выравнивании времени ответа
	dummyUUID = "00000000-0000-0000-0000-000000000000"
)

// dummyPasswordHash вычисляется один раз при старте сервиса и используется
// в Login для выравнивания времени ответа, когда запрошенный email не найден.
var dummyPasswordHash string

func init() {
	dummyPasswordHash = password.DummyHash()
}

type AuthService struct {
	db              *postgresDB.DatabaseRepository
	cache           *redisDB.CacheRepository
	publisher       messaging.Publisher
	jwtPrivateKey   *ecdsa.PrivateKey
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	appEnv          string
	pb.UnimplementedAuthServiceServer
}

func NewAuthService(db *postgresDB.DatabaseRepository, cache *redisDB.CacheRepository, publisher messaging.Publisher, jwtPrivateKey *ecdsa.PrivateKey, accessTokenTTL, refreshTokenTTL time.Duration, appEnv string) *AuthService {
	return &AuthService{
		db:              db,
		cache:           cache,
		publisher:       publisher,
		jwtPrivateKey:   jwtPrivateKey,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		appEnv:          appEnv,
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

// Register Создание нового пользователя с отправкой кода верификации на почту
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

	passwordHash, err := password.Hash(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	user := entities.User{
		UserUUID:     userUUID,
		Email:        req.GetEmail(),
		PasswordHash: passwordHash,
		FirstName:    req.GetFirstName(),
		LastName:     req.GetLastName(),
		Patronymic:   req.GetPatronymic(),
	}

	createErr := s.db.User.CreateUser(ctx, user)
	if createErr.Code != 0 {
		if createErr.Code != int(codes.AlreadyExists) {
			return nil, createErr.GRPCError()
		}

		// Email уже занят — проверяем статус существующего аккаунта
		existing, getErr := s.db.User.GetUserByEmail(ctx, entities.GetUserByEmailDTO{Email: req.GetEmail()})
		if getErr.Code != 0 {
			return nil, status.Errorf(codes.Internal, "internal error")
		}

		if existing.IsVerified {
			return nil, status.Errorf(codes.AlreadyExists, "email already registered")
		}

		// Аккаунт не верифицирован — проверяем, активен ли ещё код
		_, codeErr := s.cache.Verification.GetVerificationCode(ctx, entities.GetVerificationCodeDTO{UserUUID: existing.UserUUID})
		if codeErr.Code == 0 {
			// Код ещё активен — повторная регистрация недоступна
			return nil, status.Errorf(codes.AlreadyExists, "verification email already sent, please check your inbox")
		}

		// Код истёк — удаляем старый неверифицированный аккаунт и создаём новый
		if err := s.db.User.DeleteUser(ctx, entities.DeleteUserDTO{UserUUID: existing.UserUUID}).GRPCError(); err != nil {
			return nil, err
		}
		if err := s.db.User.CreateUser(ctx, user).GRPCError(); err != nil {
			return nil, err
		}
	}

	// Генерируем и сохраняем код верификации
	code := utils.GenerateVerificationCode()
	if err := s.cache.Verification.SaveVerificationCode(ctx, entities.SaveVerificationCodeDTO{
		UserUUID: userUUID,
		Code:     code,
	}).GRPCError(); err != nil {
		return nil, err
	}

	// Отправляем сообщение в message broker
	_ = s.publisher.SendVerificationEmail(ctx, entities.VerificationEmailMsg{
		UserUUID:  userUUID,
		Email:     req.GetEmail(),
		FirstName: req.GetFirstName(),
		Code:      code,
	})

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

	user, getErr := s.db.User.GetUserByEmail(ctx, entities.GetUserByEmailDTO{Email: req.GetEmail()})
	if getErr.Code != 0 {
		// Запускаем Argon2id на фиктивном хеше, чтобы путь "email не найден" занимал столько же времени, что и "неверный пароль".
		_, _ = password.Verify(dummyPasswordHash, req.GetPassword())
		return nil, status.Errorf(codes.InvalidArgument, "wrong email or password")
	}

	ok, err := password.Verify(user.PasswordHash, req.GetPassword())
	if err != nil || !ok {
		return nil, status.Errorf(codes.InvalidArgument, "wrong email or password")
	}

	// Проверяем, что аккаунт не удалён
	if user.DeletedAt != nil {
		hoursLeft := hoursUntilAnonymization(*user.DeletedAt)
		return nil, status.Errorf(codes.PermissionDenied,
			"account is deleted, you have %d hours to restore it", hoursLeft)
	}

	// Проверяем статус аккаунта
	if !user.IsVerified {
		return nil, status.Errorf(codes.PermissionDenied, "account not verified")
	}

	// Если у пользователя включена 2FA
	if user.Enabled2FA {
		sessionUUID := uuid.Must(uuid.NewV7()).String()
		code := utils.GenerateTwoFACode()

		// Сохраняем данные для 2FA авторизации
		if err := s.cache.TwoFA.Save2FAData(ctx, entities.Save2FADataDTO{
			SessionUUID: sessionUUID,
			UserUUID:    user.UserUUID,
			Code:        code,
		}).GRPCError(); err != nil {
			return nil, err
		}

		// Отправляем сообщение в message broker
		_ = s.publisher.Send2FAEmail(ctx, entities.TwoFAEmailMsg{
			UserUUID:  user.UserUUID,
			Email:     req.GetEmail(),
			FirstName: user.FirstName,
			Code:      code,
		})

		// Возвращаем ответ для включенной 2FA
		return &pb.LoginResponse{SessionUuid: sessionUUID}, nil
	}

	// Если 2FA выключена, создаем токены
	tokenPair, err := utils.CreateTokens(user.UserUUID, s.jwtPrivateKey, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	session := &entities.SessionInfo{}
	session.FromProto(req.GetSession())

	if err := s.cache.Auth.SaveRefreshToken(ctx, entities.SaveRefreshTokenDTO{
		UserUUID:    user.UserUUID,
		HashedToken: utils.HashToken(tokenPair.RefreshToken),
		Session:     session,
	}).GRPCError(); err != nil {
		return nil, err
	}

	// Возвращаем ответ для выключенной 2FA
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

	user, getErr := s.db.User.GetUser(ctx, entities.GetUserDTO{UserUUID: req.GetUserUuid()})
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
		DeletedAt:  format.TimePtr(user.DeletedAt),
	}, nil
}

// ChangePassword Обновление пароля пользователя с проверкой старого пароля
func (s *AuthService) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}
	if err := validate.Password(req.GetOldPassword()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid old password")
	}
	if err := validate.Password(req.GetPassword()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid new password")
	}

	// Получаем пользователя для проверки старого пароля и отправки уведомления
	user, getErr := s.db.User.GetUser(ctx, entities.GetUserDTO{UserUUID: req.GetUserUuid()})
	if getErr.Code != 0 {
		return nil, getErr.GRPCError()
	}

	// Проверяем старый пароль
	if ok, err := password.Verify(user.PasswordHash, req.GetOldPassword()); err != nil || !ok {
		return nil, status.Errorf(codes.InvalidArgument, "wrong old password")
	}

	newPasswordHash, err := password.Hash(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	if err := s.db.User.UpdateUserPassword(ctx, entities.UpdateUserPasswordDTO{
		UserUUID:     req.GetUserUuid(),
		PasswordHash: newPasswordHash,
	}).GRPCError(); err != nil {
		return nil, err
	}

	// Уведомляем пользователя о смене пароля (fire and forget)
	_ = s.publisher.SendPasswordChangedEmail(ctx, entities.PasswordChangedEmailMsg{
		UserUUID:  req.GetUserUuid(),
		Email:     user.Email,
		FirstName: user.FirstName,
	})

	// Отзываем все токены после смены пароля
	revokeErr := s.cache.Auth.RevokeAllRefreshTokens(ctx, entities.RevokeAllRefreshTokensDTO{UserUUID: req.GetUserUuid()})
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

	if err := s.db.User.UpdateUserBio(ctx, entities.UserUpdateBioDTO{
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

	// Мягкое удаление - проставляем deleted_at = NOW()
	if err := s.db.User.DeleteUser(ctx, entities.DeleteUserDTO{UserUUID: req.GetTargetUserUuid()}).GRPCError(); err != nil {
		return nil, err
	}

	// Отзываем все активные сессии - удалённый аккаунт не должен оставаться авторизованным
	revokeErr := s.cache.Auth.RevokeAllRefreshTokens(ctx, entities.RevokeAllRefreshTokensDTO{UserUUID: req.GetTargetUserUuid()})
	if revokeErr.Code != 0 && revokeErr.Code != int(codes.NotFound) {
		if err := revokeErr.GRPCError(); err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

// GetAllActiveTokens Получение всех активных токенов пользователя
func (s *AuthService) GetAllActiveTokens(ctx context.Context, req *pb.GetAllActiveTokensRequest) (*pb.GetAllActiveTokensResponse, error) {
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	userTokens, getErr := s.cache.Auth.GetAllRefreshTokens(ctx, entities.GetAllRefreshTokensDTO{UserUUID: req.GetUserUuid()})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	tokens := make([]*pb.Token, 0, len(userTokens))
	for _, entry := range userTokens {
		tokens = append(tokens, &pb.Token{
			Token:   entry.TokenHash,
			Session: entry.Session.ToProto(),
		})
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

	if err = s.cache.Auth.CheckRefreshTokenExists(ctx, entities.CheckRefreshTokenExistsDTO{
		UserUUID: tokenClaims.UserUUID,
		RawToken: req.GetRefreshToken(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	if _, getErr := s.db.User.GetUser(ctx, entities.GetUserDTO{UserUUID: tokenClaims.UserUUID}); getErr.Code != 0 {
		return nil, getErr.GRPCError()
	}

	tokenPair, err := utils.CreateTokens(tokenClaims.UserUUID, s.jwtPrivateKey, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	if err = s.cache.Auth.RefreshToken(ctx, entities.RefreshTokenDTO{
		UserUUID:     tokenClaims.UserUUID,
		OldHashToken: utils.HashToken(req.GetRefreshToken()),
		NewHashToken: utils.HashToken(tokenPair.RefreshToken),
		LastIP:       req.GetIp(),
		LastActiveAt: time.Now(),
	}).GRPCError(); err != nil {
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

	if err := s.cache.Auth.RevokeRefreshToken(ctx, entities.RevokeRefreshTokenDTO{
		UserUUID:  req.GetUserUuid(),
		TokenHash: req.GetTokenHash(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RevokeAllTokens Отзыв всех refresh токенов пользователя
func (s *AuthService) RevokeAllTokens(ctx context.Context, req *pb.RevokeAllTokensRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	if err := s.cache.Auth.RevokeAllRefreshTokens(ctx, entities.RevokeAllRefreshTokensDTO{
		UserUUID: req.GetUserUuid(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// VerifyAccount Подтверждение аккаунта по коду из письма
func (s *AuthService) VerifyAccount(ctx context.Context, req *pb.VerifyAccountRequest) (*emptypb.Empty, error) {
	if err := validate.Email(req.GetEmail()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}
	if err := validate.UserVerificationCode(req.GetCode()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid verification code format")
	}

	user, getErr := s.db.User.GetUserByEmail(ctx, entities.GetUserByEmailDTO{Email: req.GetEmail()})
	if getErr.Code != 0 {
		// Защита от timing-атаки: выполняем Redis-запрос, аналогичный тому,
		// что делается при нахождении пользователя, чтобы выровнять время ответа.
		// dummyUUID никогда не существует в Redis — запрос всегда промахивается.
		_, _ = s.cache.Verification.GetVerificationCode(ctx, entities.GetVerificationCodeDTO{UserUUID: dummyUUID})
		return nil, status.Errorf(codes.InvalidArgument, "invalid email or code")
	}

	if user.IsVerified {
		return nil, status.Errorf(codes.AlreadyExists, "account already verified")
	}

	// Проверяем наличие активного кода.
	storedCode, codeErr := s.cache.Verification.GetVerificationCode(ctx, entities.GetVerificationCodeDTO{UserUUID: user.UserUUID})
	if codeErr.Code != 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired verification code")
	}

	// Увеличиваем счётчик попыток до проверки кода
	attempts, attemptsErr := s.cache.Verification.IncrVerificationAttempts(ctx, entities.IncrVerificationAttemptsDTO{UserUUID: user.UserUUID})
	if attemptsErr.Code != 0 {
		return nil, status.Errorf(codes.Internal, "internal error")
	}
	if attempts > maxVerificationAttempts {
		_ = s.cache.Verification.DeleteVerificationCode(ctx, entities.DeleteVerificationCodeDTO{UserUUID: user.UserUUID})
		return nil, status.Errorf(codes.ResourceExhausted, "too many attempts, please request a new verification code")
	}

	if req.GetCode() != storedCode {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired verification code")
	}

	// Код верный - удаляем его из Redis
	_ = s.cache.Verification.DeleteVerificationCode(ctx, entities.DeleteVerificationCodeDTO{UserUUID: user.UserUUID})

	// Помечаем аккаунт верифицированным
	if err := s.db.User.SetUserVerified(ctx, entities.SetUserVerifiedDTO{UserUUID: user.UserUUID}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ResendVerificationCode Повторная отправка кода верификации
func (s *AuthService) ResendVerificationCode(ctx context.Context, req *pb.ResendVerificationCodeRequest) (*emptypb.Empty, error) {
	if err := validate.Email(req.GetEmail()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}

	user, getErr := s.db.User.GetUserByEmail(ctx, entities.GetUserByEmailDTO{Email: req.GetEmail()})
	if getErr.Code != 0 {
		// Не раскрываем, зарегистрирован ли email
		log.Warn().Time("time", time.Now()).Str("id", interceptors.OperationIDFromContext(ctx)).Str("method", "ResendVerificationCode").Msg("user not found")
		return &emptypb.Empty{}, nil
	}

	if user.IsVerified {
		return nil, status.Errorf(codes.AlreadyExists, "account already verified")
	}

	// Генерируем новый код (перезаписывает старый и сбрасывает счётчик попыток)
	code := utils.GenerateVerificationCode()
	if err := s.cache.Verification.SaveVerificationCode(ctx, entities.SaveVerificationCodeDTO{
		UserUUID: user.UserUUID,
		Code:     code,
	}).GRPCError(); err != nil {
		return nil, err
	}

	// Отправляем сообщение в message broker
	_ = s.publisher.SendVerificationEmail(ctx, entities.VerificationEmailMsg{
		UserUUID:  user.UserUUID,
		Email:     req.GetEmail(),
		FirstName: user.FirstName,
		Code:      code,
	})

	return &emptypb.Empty{}, nil
}

// ForgotPassword Запрашивает восстановление пароля — отправляет код на почту.
func (s *AuthService) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*emptypb.Empty, error) {
	if err := validate.Email(req.GetEmail()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}

	user, getErr := s.db.User.GetUserByEmail(ctx, entities.GetUserByEmailDTO{Email: req.GetEmail()})
	// Не раскрываем, зарегистрирован ли email
	if getErr.Code != 0 {
		log.Warn().Time("time", time.Now()).Str("id", interceptors.OperationIDFromContext(ctx)).Str("method", "ForgotPassword").Msg("user not found")
		return &emptypb.Empty{}, nil
	}

	// Не раскрываем состояние аккаунта
	if !user.IsVerified {
		log.Warn().Time("time", time.Now()).Str("id", interceptors.OperationIDFromContext(ctx)).Str("method", "ForgotPassword").Msg("user not verified")
		return &emptypb.Empty{}, nil
	}

	code := utils.GenerateRecoveryCode()
	if err := s.cache.Recovery.SaveRecoveryCode(ctx, entities.SaveRecoveryCodeDTO{
		UserUUID: user.UserUUID,
		Code:     code,
	}).GRPCError(); err != nil {
		return nil, err
	}

	// Отправляем сообщение в message broker
	_ = s.publisher.SendRecoveryEmail(ctx, entities.RecoveryEmailMsg{
		UserUUID:  user.UserUUID,
		Email:     req.GetEmail(),
		FirstName: user.FirstName,
		Code:      code,
	})

	return &emptypb.Empty{}, nil
}

// ResetPassword Сбрасывает пароль по коду восстановления
func (s *AuthService) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*emptypb.Empty, error) {
	if err := validate.Email(req.GetEmail()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}
	if err := validate.UserRecoveryCode(req.GetCode()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid code format")
	}
	if err := validate.Password(req.GetNewPassword()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}

	user, getErr := s.db.User.GetUserByEmail(ctx, entities.GetUserByEmailDTO{Email: req.GetEmail()})
	if getErr.Code != 0 {
		// Защита от timing-атаки: выполняем Redis-запрос, аналогичный тому,
		// что делается при нахождении пользователя, чтобы выровнять время ответа.
		// dummyUUID никогда не существует в Redis — запрос всегда промахивается.
		_, _ = s.cache.Recovery.GetRecoveryCode(ctx, entities.GetRecoveryCodeDTO{UserUUID: dummyUUID})
		return nil, status.Errorf(codes.InvalidArgument, "invalid email or code")
	}

	if !user.IsVerified {
		// Аналогично: выравниваем время ответа для не верифицированного пользователя.
		_, _ = s.cache.Recovery.GetRecoveryCode(ctx, entities.GetRecoveryCodeDTO{UserUUID: dummyUUID})
		return nil, status.Errorf(codes.InvalidArgument, "invalid email or code")
	}

	storedCode, codeErr := s.cache.Recovery.GetRecoveryCode(ctx, entities.GetRecoveryCodeDTO{UserUUID: user.UserUUID})
	if codeErr.Code != 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired code")
	}

	attempts, attemptsErr := s.cache.Recovery.IncrRecoveryAttempts(ctx, entities.IncrRecoveryAttemptsDTO{UserUUID: user.UserUUID})
	if attemptsErr.Code != 0 {
		return nil, status.Errorf(codes.Internal, "internal error")
	}
	if attempts > maxRecoveryAttempts {
		_ = s.cache.Recovery.DeleteRecoveryCode(ctx, entities.DeleteRecoveryCodeDTO{UserUUID: user.UserUUID})
		return nil, status.Errorf(codes.ResourceExhausted, "too many attempts, please request a new recovery code")
	}

	if req.GetCode() != storedCode {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired code")
	}

	_ = s.cache.Recovery.DeleteRecoveryCode(ctx, entities.DeleteRecoveryCodeDTO{UserUUID: user.UserUUID})

	passwordHash, err := password.Hash(req.GetNewPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	if err = s.db.User.UpdateUserPassword(ctx, entities.UpdateUserPasswordDTO{
		UserUUID:     user.UserUUID,
		PasswordHash: passwordHash,
	}).GRPCError(); err != nil {
		return nil, err
	}

	// Отзываем все активные сессии после смены пароля
	revokeErr := s.cache.Auth.RevokeAllRefreshTokens(ctx, entities.RevokeAllRefreshTokensDTO{UserUUID: user.UserUUID})
	if revokeErr.Code != 0 && revokeErr.Code != int(codes.NotFound) {
		if err = revokeErr.GRPCError(); err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

// Verify2FA Подтверждение 2FA авторизации
func (s *AuthService) Verify2FA(ctx context.Context, req *pb.Verify2FARequest) (*pb.Verify2FAResponse, error) {
	// Валидации
	if err := validate.UUID(req.SessionUuid); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid session uuid")
	}
	if err := validate.User2FACCode(req.GetCode()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid code format")
	}

	// Получаем данные по sessionUUID
	data, getErr := s.cache.TwoFA.Get2FAData(ctx, entities.Get2FADataDTO{SessionUUID: req.GetSessionUuid()})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	// Увеличиваем счётчик попыток
	attempts, attemptsErr := s.cache.TwoFA.Incr2FAAttempts(ctx, entities.Incr2FAAttemptsDTO{SessionUUID: req.GetSessionUuid()})
	if err := attemptsErr.GRPCError(); err != nil {
		return nil, err
	}
	if attempts > max2FAAttempts {
		_ = s.cache.TwoFA.Delete2FAData(ctx, entities.Delete2FADataDTO{SessionUUID: req.GetSessionUuid()})
		return nil, status.Errorf(codes.ResourceExhausted, "too many attempts, please try login again")
	}

	// Сравниваем code
	if data.Code != req.GetCode() {
		return nil, status.Errorf(codes.PermissionDenied, "invalid or expired code")
	}

	// Удаляем сессию
	_ = s.cache.TwoFA.Delete2FAData(ctx, entities.Delete2FADataDTO{SessionUUID: req.GetSessionUuid()})

	// Генерируем пару токенов
	tokenPair, err := utils.CreateTokens(data.UserUUID, s.jwtPrivateKey, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	session := &entities.SessionInfo{}
	session.FromProto(req.GetSession())

	// Сохраняем refresh токен
	if err := s.cache.Auth.SaveRefreshToken(ctx, entities.SaveRefreshTokenDTO{
		UserUUID:    data.UserUUID,
		HashedToken: utils.HashToken(tokenPair.RefreshToken),
		Session:     session,
	}).GRPCError(); err != nil {
		return nil, err
	}

	// Возвращаем токены и userUUID
	return &pb.Verify2FAResponse{
		UserUuid:     data.UserUUID,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// UpdateUser2FA Включение / выключение 2FA авторизации
func (s *AuthService) UpdateUser2FA(ctx context.Context, req *pb.UpdateUser2FARequest) (*emptypb.Empty, error) {
	// Валидации
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	// Обновляем данные пользователя
	if err := s.db.User.UpdateUser2FA(ctx, entities.UpdateUser2FADTO{
		UserUUID:     req.GetUserUuid(),
		TwoFAEnabled: req.GetEnable_2Fa(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RestoreAccount Восстанавливает мягко удалённый аккаунт в течение 30 дней после удаления
func (s *AuthService) RestoreAccount(ctx context.Context, req *pb.RestoreAccountRequest) (*emptypb.Empty, error) {
	if err := validate.Email(req.GetEmail()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}
	if err := validate.Password(req.GetPassword()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}

	user, getErr := s.db.User.GetUserByEmail(ctx, entities.GetUserByEmailDTO{Email: req.GetEmail()})
	if getErr.Code != 0 {
		_, _ = password.Verify(dummyPasswordHash, req.GetPassword())
		return nil, status.Errorf(codes.InvalidArgument, "wrong email or password")
	}

	ok, err := password.Verify(user.PasswordHash, req.GetPassword())
	if err != nil || !ok {
		return nil, status.Errorf(codes.InvalidArgument, "wrong email or password")
	}

	if user.DeletedAt == nil {
		return nil, status.Errorf(codes.InvalidArgument, "account is not deleted")
	}

	if time.Since(*user.DeletedAt) > accountDeletionRetention {
		return nil, status.Errorf(codes.PermissionDenied, "restoration period has expired")
	}

	if err := s.db.User.RestoreUser(ctx, entities.RestoreUserDTO{UserUUID: user.UserUUID}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ

// GetVerificationCode Отладочный метод — возвращает активный код верификации.
// Доступен только при APP_ENV=test; в production возвращает Unimplemented.
func (s *AuthService) GetVerificationCode(ctx context.Context, req *pb.GetVerificationCodeRequest) (*pb.GetVerificationCodeResponse, error) {
	if s.appEnv != "test" {
		return nil, status.Errorf(codes.Unimplemented, "not available")
	}

	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	code, codeErr := s.cache.Verification.GetVerificationCode(ctx, entities.GetVerificationCodeDTO{UserUUID: req.GetUserUuid()})
	if err := codeErr.GRPCError(); err != nil {
		return nil, err
	}

	return &pb.GetVerificationCodeResponse{Code: code}, nil
}

// GetRecoveryCode Отладочный метод — возвращает активный код восстановления пароля.
// Доступен только при APP_ENV=test; в production возвращает Unimplemented.
func (s *AuthService) GetRecoveryCode(ctx context.Context, req *pb.GetRecoveryCodeRequest) (*pb.GetRecoveryCodeResponse, error) {
	if s.appEnv != "test" {
		return nil, status.Errorf(codes.Unimplemented, "not available")
	}

	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	code, codeErr := s.cache.Recovery.GetRecoveryCode(ctx, entities.GetRecoveryCodeDTO{UserUUID: req.GetUserUuid()})
	if err := codeErr.GRPCError(); err != nil {
		return nil, err
	}

	return &pb.GetRecoveryCodeResponse{Code: code}, nil
}

// Get2FACode Отладочный метод — возвращает активный код 2FA по uuid сессии.
// Доступен только при APP_ENV=test; в production возвращает Unimplemented.
func (s *AuthService) Get2FACode(ctx context.Context, req *pb.Get2FACodeRequest) (*pb.Get2FACodeResponse, error) {
	if s.appEnv != "test" {
		return nil, status.Errorf(codes.Unimplemented, "not available")
	}

	if err := validate.UUID(req.GetSessionUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid session uuid")
	}

	data, dataErr := s.cache.TwoFA.Get2FAData(ctx, entities.Get2FADataDTO{SessionUUID: req.GetSessionUuid()})
	if err := dataErr.GRPCError(); err != nil {
		return nil, err
	}

	return &pb.Get2FACodeResponse{Code: data.Code}, nil
}
