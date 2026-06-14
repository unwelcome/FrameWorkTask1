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
	max2FAAttempts        = 5
	max2FAEmailDailyCount = 10

	maxVerificationEmailDailyCount = 5
	maxRecoveryEmailDailyCount     = 3

	// accountDeletionRetention — срок хранения данных мягко удалённого аккаунта
	accountDeletionRetention = 30 * 24 * time.Hour

	// resetPasswordTokenTTL — время жизни JWT токена для сброса пароля
	resetPasswordTokenTTL = 15 * time.Minute

	// verificationTokenTTL — время жизни JWT токена для верификации аккаунта
	verificationTokenTTL = 48 * time.Hour
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
func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*emptypb.Empty, error) {
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

		// Email уже занят, проверяем, верифицирован ли он
		existing, getErr := s.db.User.GetUserByEmail(ctx, entities.GetUserByEmailDTO{Email: req.GetEmail()})
		if err := getErr.GRPCError(); err != nil {
			return nil, err
		}

		if existing.IsVerified {
			// Уведомляем владельца почты о попытке регистрации
			_ = s.publisher.SendRegistrationAttemptEmail(ctx, entities.RegistrationAttemptEmailMsg{
				Email:     existing.Email,
				FirstName: existing.FirstName,
			})
		} else {
			// Email занят неверифицированным аккаунтом — отправляем новый токен верификации
			verificationToken, tokenErr := utils.CreateVerificationToken(existing.Email, s.jwtPrivateKey, verificationTokenTTL)
			if tokenErr != nil {
				return nil, status.Errorf(codes.Internal, "internal error")
			}
			_ = s.publisher.SendVerificationEmail(ctx, entities.VerificationEmailMsg{
				UserUUID:  existing.UserUUID,
				Email:     existing.Email,
				FirstName: existing.FirstName,
				Token:     verificationToken,
			})
		}

		// Не раскрываем, что email уже зарегистрирован
		log.Warn().Time("time", time.Now()).Str("id", interceptors.OperationIDFromContext(ctx)).Str("method", "Register").Bool("verified", existing.IsVerified).Msg("user already exists")
		return &emptypb.Empty{}, nil
	}

	// Генерируем JWT токен верификации
	verificationToken, err := utils.CreateVerificationToken(req.GetEmail(), s.jwtPrivateKey, verificationTokenTTL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Отправляем сообщение в message broker
	_ = s.publisher.SendVerificationEmail(ctx, entities.VerificationEmailMsg{
		UserUUID:  userUUID,
		Email:     req.GetEmail(),
		FirstName: req.GetFirstName(),
		Token:     verificationToken,
	})

	return &emptypb.Empty{}, nil
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
		return nil, status.Error(codes.PermissionDenied, deletedAccountMessage(*user.DeletedAt))
	}

	// Проверяем статус аккаунта
	if !user.IsVerified {
		return nil, status.Errorf(codes.PermissionDenied, "account is not verified")
	}

	// Если у пользователя включена 2FA
	if user.Enabled2FA {
		// Rate limiting: cooldown между отправками 2FA писем
		allowed, rateLimitErr := s.cache.TwoFA.Acquire2FAEmailCooldown(ctx, entities.Acquire2FAEmailCooldownDTO{UserUUID: user.UserUUID})
		if err := rateLimitErr.GRPCError(); err != nil {
			return nil, err
		}
		if !allowed {
			return nil, status.Errorf(codes.ResourceExhausted, "please wait before requesting a new 2FA code")
		}

		// Rate limiting: суточный лимит отправок 2FA писем
		count, countErr := s.cache.TwoFA.Incr2FAEmailDailyCount(ctx, entities.Incr2FAEmailDailyCountDTO{UserUUID: user.UserUUID})
		if err := countErr.GRPCError(); err != nil {
			return nil, err
		}
		if count > max2FAEmailDailyCount {
			return nil, status.Errorf(codes.ResourceExhausted, "daily 2FA email limit reached")
		}

		sessionUUID := uuid.Must(uuid.NewV7()).String()
		code := utils.GenerateTwoFACode()

		// Сохраняем данные для 2FA авторизации
		if err := s.cache.TwoFA.Save2FAData(ctx, entities.Save2FADataDTO{
			SessionUUID: sessionUUID,
			UserUUID:    user.UserUUID,
			Email:       user.Email,
			FirstName:   user.FirstName,
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

	if err := s.cache.Auth.SaveSession(ctx, entities.SaveSessionDTO{
		UserUUID:    user.UserUUID,
		HashedToken: utils.HashToken(tokenPair.RefreshToken),
		Session:     session,
	}).GRPCError(); err != nil {
		return nil, err
	}

	// Уведомляем пользователя об успешном входе
	_ = s.publisher.SendLoginNotificationEmail(ctx, entities.LoginNotificationEmailMsg{
		UserUUID:  user.UserUUID,
		Email:     user.Email,
		FirstName: user.FirstName,
		IP:        session.IP,
		Browser:   session.Browser,
		OS:        session.OS,
		LoginAt:   time.Now().Unix(),
	})

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
		UserUuid:    user.UserUUID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Patronymic:  user.Patronymic,
		Description: user.Description,
		CreatedAt:   user.CreatedAt,
		DeletedAt:   format.TimePtr(user.DeletedAt),
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

	if user.DeletedAt != nil {
		return nil, status.Error(codes.PermissionDenied, deletedAccountMessage(*user.DeletedAt))
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
	revokeErr := s.cache.Auth.RevokeAllSessions(ctx, entities.RevokeAllSessionsDTO{UserUUID: req.GetUserUuid()})
	if revokeErr.Code != 0 && revokeErr.Code != int(codes.NotFound) {
		if err := revokeErr.GRPCError(); err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

// UpdateUserBio Обновление ФИО и описания пользователя
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
	if err := validate.UserDescription(req.GetDescription()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid description")
	}

	user, getErr := s.db.User.GetUser(ctx, entities.GetUserDTO{UserUUID: req.GetUserUuid()})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}
	if user.DeletedAt != nil {
		return nil, status.Error(codes.PermissionDenied, deletedAccountMessage(*user.DeletedAt))
	}

	if err := s.db.User.UpdateUserBio(ctx, entities.UserUpdateBioDTO{
		UserUUID:    req.GetUserUuid(),
		FirstName:   req.GetFirstName(),
		LastName:    req.GetLastName(),
		Patronymic:  req.GetPatronymic(),
		Description: req.GetDescription(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// DeleteUser Удаление пользователя
func (s *AuthService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	// Мягкое удаление - проставляем deleted_at = NOW()
	if err := s.db.User.DeleteUser(ctx, entities.DeleteUserDTO{UserUUID: req.GetUserUuid()}).GRPCError(); err != nil {
		return nil, err
	}

	// Отзываем все активные сессии - удалённый аккаунт не должен оставаться авторизованным
	revokeErr := s.cache.Auth.RevokeAllSessions(ctx, entities.RevokeAllSessionsDTO{UserUUID: req.GetUserUuid()})
	if revokeErr.Code != 0 && revokeErr.Code != int(codes.NotFound) {
		if err := revokeErr.GRPCError(); err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

// GetAllActiveSessions Получение всех активных сессий пользователя
func (s *AuthService) GetAllActiveSessions(ctx context.Context, req *pb.GetAllActiveSessionsRequest) (*pb.GetAllActiveSessionsResponse, error) {
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	userTokens, getErr := s.cache.Auth.GetAllSessions(ctx, entities.GetAllSessionsDTO{UserUUID: req.GetUserUuid()})
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

	return &pb.GetAllActiveSessionsResponse{Tokens: tokens}, nil
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

	if err = s.cache.Auth.CheckSessionExists(ctx, entities.CheckSessionExistsDTO{
		UserUUID: tokenClaims.UserUUID,
		RawToken: req.GetRefreshToken(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	user, getErr := s.db.User.GetUser(ctx, entities.GetUserDTO{UserUUID: tokenClaims.UserUUID})
	if getErr.Code != 0 {
		return nil, getErr.GRPCError()
	}
	if user.DeletedAt != nil {
		return nil, status.Error(codes.PermissionDenied, "account deleted")
	}
	if !user.IsVerified {
		return nil, status.Error(codes.PermissionDenied, "account not verified")
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

// RevokeSession Отзыв сессии пользователя по хешу refresh токена
func (s *AuthService) RevokeSession(ctx context.Context, req *pb.RevokeSessionRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}
	if strings.TrimSpace(req.GetTokenHash()) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "token hash missed")
	}

	if err := s.cache.Auth.RevokeSession(ctx, entities.RevokeSessionDTO{
		UserUUID:  req.GetUserUuid(),
		TokenHash: req.GetTokenHash(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RevokeAllSessions Отзыв всех сессий пользователя
func (s *AuthService) RevokeAllSessions(ctx context.Context, req *pb.RevokeAllSessionsRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetUserUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	revokeErr := s.cache.Auth.RevokeAllSessions(ctx, entities.RevokeAllSessionsDTO{
		UserUUID: req.GetUserUuid(),
	})
	if revokeErr.Code != 0 && revokeErr.Code != int(codes.NotFound) {
		if err := revokeErr.GRPCError(); err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

// VerifyAccount Подтверждение аккаунта по JWT токену из письма (magic link)
func (s *AuthService) VerifyAccount(ctx context.Context, req *pb.VerifyAccountRequest) (*emptypb.Empty, error) {
	claims, err := utils.ParseVerificationToken(req.GetVerificationToken(), s.jwtPrivateKey)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired verification token")
	}

	if claims.TokenType != entities.VerificationTokenType {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired verification token")
	}

	// Проверяем, что токен не был использован ранее
	blacklisted, blacklistErr := s.cache.Verification.IsVerificationTokenBlacklisted(ctx, entities.IsVerificationTokenBlacklistedDTO{TokenID: claims.ID})
	if err := blacklistErr.GRPCError(); err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired verification token")
	}

	user, getErr := s.db.User.GetUserByEmail(ctx, entities.GetUserByEmailDTO{Email: claims.Email})
	if getErr.Code != 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired verification token")
	}

	if user.IsVerified {
		return nil, status.Errorf(codes.AlreadyExists, "account already verified")
	}

	// Помечаем аккаунт верифицированным
	if err := s.db.User.SetUserVerified(ctx, entities.SetUserVerifiedDTO{UserUUID: user.UserUUID}).GRPCError(); err != nil {
		return nil, err
	}

	// Добавляем токен в blacklist — одноразовое использование
	remainingTTL := time.Until(claims.ExpiresAt.Time)
	if remainingTTL > 0 {
		_ = s.cache.Verification.AddToVerificationTokenBlacklist(ctx, entities.AddToVerificationTokenBlacklistDTO{
			TokenID: claims.ID,
			TTL:     remainingTTL,
		})
	}

	return &emptypb.Empty{}, nil
}

// ResendVerificationCode Повторная отправка письма верификации (magic link)
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
		log.Warn().Time("time", time.Now()).Str("id", interceptors.OperationIDFromContext(ctx)).Str("method", "ResendVerificationCode").Msg("user already verified")
		return &emptypb.Empty{}, nil
	}

	// Rate limiting: cooldown между отправками писем верификации
	allowed, rateLimitErr := s.cache.Verification.AcquireVerificationEmailCooldown(ctx, entities.AcquireVerificationEmailCooldownDTO{UserUUID: user.UserUUID})
	if err := rateLimitErr.GRPCError(); err != nil {
		return nil, err
	}
	if !allowed {
		log.Warn().Time("time", time.Now()).Str("id", interceptors.OperationIDFromContext(ctx)).Str("method", "ResendVerificationCode").Msg("verification email cooldown active")
		return &emptypb.Empty{}, nil
	}

	// Rate limiting: суточный лимит отправок писем верификации
	count, countErr := s.cache.Verification.IncrVerificationEmailDailyCount(ctx, entities.IncrVerificationEmailDailyCountDTO{UserUUID: user.UserUUID})
	if err := countErr.GRPCError(); err != nil {
		return nil, err
	}
	if count > maxVerificationEmailDailyCount {
		log.Warn().Time("time", time.Now()).Str("id", interceptors.OperationIDFromContext(ctx)).Str("method", "ResendVerificationCode").Msg("verification email daily limit reached")
		return &emptypb.Empty{}, nil
	}

	// Генерируем новый JWT токен верификации
	verificationToken, err := utils.CreateVerificationToken(user.Email, s.jwtPrivateKey, verificationTokenTTL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Отправляем сообщение в message broker
	_ = s.publisher.SendVerificationEmail(ctx, entities.VerificationEmailMsg{
		UserUUID:  user.UserUUID,
		Email:     user.Email,
		FirstName: user.FirstName,
		Token:     verificationToken,
	})

	return &emptypb.Empty{}, nil
}

// ForgotPassword Запрос восстановления пароля
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
	if user.DeletedAt != nil {
		log.Warn().Time("time", time.Now()).Str("id", interceptors.OperationIDFromContext(ctx)).Str("method", "ForgotPassword").Msg("account is deleted")
		return &emptypb.Empty{}, nil
	}

	// Rate limiting: cooldown между отправками писем восстановления пароля
	allowed, rateLimitErr := s.cache.Recovery.AcquireRecoveryEmailCooldown(ctx, entities.AcquireRecoveryEmailCooldownDTO{UserUUID: user.UserUUID})
	if err := rateLimitErr.GRPCError(); err != nil {
		return nil, err
	}
	if !allowed {
		log.Warn().Time("time", time.Now()).Str("id", interceptors.OperationIDFromContext(ctx)).Str("method", "ForgotPassword").Msg("recovery email cooldown active")
		return &emptypb.Empty{}, nil
	}

	// Rate limiting: суточный лимит отправок писем восстановления пароля
	count, countErr := s.cache.Recovery.IncrRecoveryEmailDailyCount(ctx, entities.IncrRecoveryEmailDailyCountDTO{UserUUID: user.UserUUID})
	if err := countErr.GRPCError(); err != nil {
		return nil, err
	}
	if count > maxRecoveryEmailDailyCount {
		log.Warn().Time("time", time.Now()).Str("id", interceptors.OperationIDFromContext(ctx)).Str("method", "ForgotPassword").Msg("recovery email daily limit reached")
		return &emptypb.Empty{}, nil
	}

	resetToken, err := utils.CreateResetPasswordToken(user.Email, s.jwtPrivateKey, resetPasswordTokenTTL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Отправляем сообщение в message broker
	_ = s.publisher.SendRecoveryEmail(ctx, entities.RecoveryEmailMsg{
		UserUUID:  user.UserUUID,
		Email:     user.Email,
		FirstName: user.FirstName,
		Token:     resetToken,
	})

	return &emptypb.Empty{}, nil
}

// ResetPassword Сбрасывает пароль по JWT reset-password токену
func (s *AuthService) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*emptypb.Empty, error) {
	if err := validate.Password(req.GetNewPassword()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}

	// Верифицируем JWT токен (подпись + срок действия)
	claims, err := utils.ParseResetPasswordToken(req.GetResetToken(), s.jwtPrivateKey)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired reset token")
	}

	if claims.TokenType != entities.ResetPasswordTokenType {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired reset token")
	}

	// Проверяем, что токен не был использован ранее
	blacklisted, blacklistErr := s.cache.Recovery.IsResetTokenBlacklisted(ctx, entities.IsResetTokenBlacklistedDTO{TokenID: claims.ID})
	if err := blacklistErr.GRPCError(); err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired reset token")
	}

	user, getErr := s.db.User.GetUserByEmail(ctx, entities.GetUserByEmailDTO{Email: claims.Email})
	if getErr.Code != 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired reset token")
	}

	if !user.IsVerified || user.DeletedAt != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired reset token")
	}

	passwordHash, hashErr := password.Hash(req.GetNewPassword())
	if hashErr != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	if err = s.db.User.UpdateUserPassword(ctx, entities.UpdateUserPasswordDTO{
		UserUUID:     user.UserUUID,
		PasswordHash: passwordHash,
	}).GRPCError(); err != nil {
		return nil, err
	}

	// Добавляем токен в blacklist с TTL равным оставшемуся времени жизни
	remainingTTL := time.Until(claims.ExpiresAt.Time)
	if remainingTTL > 0 {
		_ = s.cache.Recovery.AddToResetTokenBlacklist(ctx, entities.AddToResetTokenBlacklistDTO{
			TokenID: claims.ID,
			TTL:     remainingTTL,
		})
	}

	// Уведомляем пользователя о сбросе пароля
	_ = s.publisher.SendPasswordResetEmail(ctx, entities.PasswordResetEmailMsg{
		UserUUID:  user.UserUUID,
		Email:     user.Email,
		FirstName: user.FirstName,
	})

	// Отзываем все активные сессии после смены пароля
	revokeErr := s.cache.Auth.RevokeAllSessions(ctx, entities.RevokeAllSessionsDTO{UserUUID: user.UserUUID})
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired code")
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
	if err := s.cache.Auth.SaveSession(ctx, entities.SaveSessionDTO{
		UserUUID:    data.UserUUID,
		HashedToken: utils.HashToken(tokenPair.RefreshToken),
		Session:     session,
	}).GRPCError(); err != nil {
		return nil, err
	}

	// Уведомляем пользователя об успешном входе
	_ = s.publisher.SendLoginNotificationEmail(ctx, entities.LoginNotificationEmailMsg{
		UserUUID:  data.UserUUID,
		Email:     data.Email,
		FirstName: data.FirstName,
		IP:        session.IP,
		Browser:   session.Browser,
		OS:        session.OS,
		LoginAt:   time.Now().Unix(),
	})

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

	user, getErr := s.db.User.GetUser(ctx, entities.GetUserDTO{UserUUID: req.GetUserUuid()})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}
	if user.DeletedAt != nil {
		return nil, status.Error(codes.PermissionDenied, deletedAccountMessage(*user.DeletedAt))
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

	anonymizationTime := nextCleanupTimeAfter(user.DeletedAt.Add(accountDeletionRetention))
	if !time.Now().Before(anonymizationTime) {
		return nil, status.Errorf(codes.PermissionDenied, "restoration period has expired")
	}

	if err := s.db.User.RestoreUser(ctx, entities.RestoreUserDTO{UserUUID: user.UserUUID}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ─── Вспомогательные функции ──────────────────────────────────────────────────

// GetVerificationToken Отладочный метод — генерирует и возвращает токен верификации по email.
// Доступен только при APP_ENV=test; в production возвращает Unimplemented.
func (s *AuthService) GetVerificationToken(ctx context.Context, req *pb.GetVerificationTokenRequest) (*pb.GetVerificationTokenResponse, error) {
	if s.appEnv != "test" {
		return nil, status.Errorf(codes.Unimplemented, "not available")
	}

	if err := validate.Email(req.GetEmail()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}

	user, getErr := s.db.User.GetUserByEmail(ctx, entities.GetUserByEmailDTO{Email: req.GetEmail()})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	token, err := utils.CreateVerificationToken(user.Email, s.jwtPrivateKey, verificationTokenTTL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &pb.GetVerificationTokenResponse{Token: token}, nil
}

// GetResetPasswordToken Отладочный метод — генерирует и возвращает reset-password токен по email.
// Доступен только при APP_ENV=test; в production возвращает Unimplemented.
func (s *AuthService) GetResetPasswordToken(ctx context.Context, req *pb.GetResetPasswordTokenRequest) (*pb.GetResetPasswordTokenResponse, error) {
	if s.appEnv != "test" {
		return nil, status.Errorf(codes.Unimplemented, "not available")
	}

	if err := validate.Email(req.GetEmail()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}

	user, getErr := s.db.User.GetUserByEmail(ctx, entities.GetUserByEmailDTO{Email: req.GetEmail()})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	token, err := utils.CreateResetPasswordToken(user.Email, s.jwtPrivateKey, resetPasswordTokenTTL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &pb.GetResetPasswordTokenResponse{Token: token}, nil
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
