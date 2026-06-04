package services

import (
	"context"
	"crypto/ecdsa"
	"time"

	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/messaging"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/utils"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
)

// ─── Mock: UserRepository ────────────────────────────────────────────────────

type mockUserRepo struct {
	createUser         func(ctx context.Context, dto entities.User) Error.CodeError
	getUserByEmail     func(ctx context.Context, dto entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError)
	getUser            func(ctx context.Context, dto entities.GetUserDTO) (*entities.UserGet, Error.CodeError)
	updateUserPassword func(ctx context.Context, dto entities.UpdateUserPasswordDTO) Error.CodeError
	updateUserBio      func(ctx context.Context, dto entities.UserUpdateBio) Error.CodeError
	deleteUser         func(ctx context.Context, dto entities.DeleteUserDTO) Error.CodeError
	setUserVerified    func(ctx context.Context, dto entities.SetUserVerifiedDTO) Error.CodeError
}

func (m *mockUserRepo) CreateUser(ctx context.Context, dto entities.User) Error.CodeError {
	return m.createUser(ctx, dto)
}
func (m *mockUserRepo) GetUserByEmail(ctx context.Context, dto entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
	return m.getUserByEmail(ctx, dto)
}
func (m *mockUserRepo) GetUser(ctx context.Context, dto entities.GetUserDTO) (*entities.UserGet, Error.CodeError) {
	return m.getUser(ctx, dto)
}
func (m *mockUserRepo) UpdateUserPassword(ctx context.Context, dto entities.UpdateUserPasswordDTO) Error.CodeError {
	return m.updateUserPassword(ctx, dto)
}
func (m *mockUserRepo) UpdateUserBio(ctx context.Context, dto entities.UserUpdateBio) Error.CodeError {
	return m.updateUserBio(ctx, dto)
}
func (m *mockUserRepo) DeleteUser(ctx context.Context, dto entities.DeleteUserDTO) Error.CodeError {
	return m.deleteUser(ctx, dto)
}
func (m *mockUserRepo) SetUserVerified(ctx context.Context, dto entities.SetUserVerifiedDTO) Error.CodeError {
	return m.setUserVerified(ctx, dto)
}

// ─── Mock: AuthRepository ────────────────────────────────────────────────────

type mockAuthRepo struct {
	saveRefreshToken        func(ctx context.Context, dto entities.SaveRefreshTokenDTO) Error.CodeError
	getAllRefreshTokens     func(ctx context.Context, dto entities.GetAllRefreshTokensDTO) ([]string, Error.CodeError)
	checkRefreshTokenExists func(ctx context.Context, dto entities.CheckRefreshTokenExistsDTO) Error.CodeError
	revokeRefreshToken      func(ctx context.Context, dto entities.RevokeRefreshTokenDTO) Error.CodeError
	revokeAllRefreshTokens  func(ctx context.Context, dto entities.RevokeAllRefreshTokensDTO) Error.CodeError
	refreshToken            func(ctx context.Context, dto entities.RefreshTokenDTO) Error.CodeError
}

func (m *mockAuthRepo) SaveRefreshToken(ctx context.Context, dto entities.SaveRefreshTokenDTO) Error.CodeError {
	return m.saveRefreshToken(ctx, dto)
}
func (m *mockAuthRepo) GetAllRefreshTokens(ctx context.Context, dto entities.GetAllRefreshTokensDTO) ([]string, Error.CodeError) {
	return m.getAllRefreshTokens(ctx, dto)
}
func (m *mockAuthRepo) CheckRefreshTokenExists(ctx context.Context, dto entities.CheckRefreshTokenExistsDTO) Error.CodeError {
	return m.checkRefreshTokenExists(ctx, dto)
}
func (m *mockAuthRepo) RevokeRefreshToken(ctx context.Context, dto entities.RevokeRefreshTokenDTO) Error.CodeError {
	return m.revokeRefreshToken(ctx, dto)
}
func (m *mockAuthRepo) RevokeAllRefreshTokens(ctx context.Context, dto entities.RevokeAllRefreshTokensDTO) Error.CodeError {
	return m.revokeAllRefreshTokens(ctx, dto)
}
func (m *mockAuthRepo) RefreshToken(ctx context.Context, dto entities.RefreshTokenDTO) Error.CodeError {
	return m.refreshToken(ctx, dto)
}

// ─── Mock: VerificationRepository ───────────────────────────────────────────

type mockVerificationRepo struct {
	saveVerificationCode     func(ctx context.Context, dto entities.SaveVerificationCodeDTO) Error.CodeError
	getVerificationCode      func(ctx context.Context, dto entities.GetVerificationCodeDTO) (string, Error.CodeError)
	deleteVerificationCode   func(ctx context.Context, dto entities.DeleteVerificationCodeDTO) Error.CodeError
	incrVerificationAttempts func(ctx context.Context, dto entities.IncrVerificationAttemptsDTO) (int64, Error.CodeError)

	saveRecoveryCode     func(ctx context.Context, dto entities.SaveRecoveryCodeDTO) Error.CodeError
	getRecoveryCode      func(ctx context.Context, dto entities.GetRecoveryCodeDTO) (string, Error.CodeError)
	deleteRecoveryCode   func(ctx context.Context, dto entities.DeleteRecoveryCodeDTO) Error.CodeError
	incrRecoveryAttempts func(ctx context.Context, dto entities.IncrRecoveryAttemptsDTO) (int64, Error.CodeError)
}

func (m *mockVerificationRepo) SaveVerificationCode(ctx context.Context, dto entities.SaveVerificationCodeDTO) Error.CodeError {
	return m.saveVerificationCode(ctx, dto)
}
func (m *mockVerificationRepo) GetVerificationCode(ctx context.Context, dto entities.GetVerificationCodeDTO) (string, Error.CodeError) {
	return m.getVerificationCode(ctx, dto)
}
func (m *mockVerificationRepo) DeleteVerificationCode(ctx context.Context, dto entities.DeleteVerificationCodeDTO) Error.CodeError {
	return m.deleteVerificationCode(ctx, dto)
}
func (m *mockVerificationRepo) IncrVerificationAttempts(ctx context.Context, dto entities.IncrVerificationAttemptsDTO) (int64, Error.CodeError) {
	return m.incrVerificationAttempts(ctx, dto)
}
func (m *mockVerificationRepo) SaveRecoveryCode(ctx context.Context, dto entities.SaveRecoveryCodeDTO) Error.CodeError {
	return m.saveRecoveryCode(ctx, dto)
}
func (m *mockVerificationRepo) GetRecoveryCode(ctx context.Context, dto entities.GetRecoveryCodeDTO) (string, Error.CodeError) {
	return m.getRecoveryCode(ctx, dto)
}
func (m *mockVerificationRepo) DeleteRecoveryCode(ctx context.Context, dto entities.DeleteRecoveryCodeDTO) Error.CodeError {
	return m.deleteRecoveryCode(ctx, dto)
}
func (m *mockVerificationRepo) IncrRecoveryAttempts(ctx context.Context, dto entities.IncrRecoveryAttemptsDTO) (int64, Error.CodeError) {
	return m.incrRecoveryAttempts(ctx, dto)
}

// ─── Mock: Publisher ─────────────────────────────────────────────────────────

type mockPublisher struct {
	sendVerificationEmail func(ctx context.Context, dto entities.VerificationEmailMsg) Error.CodeError
	sendRecoveryEmail     func(ctx context.Context, dto entities.RecoveryEmailMsg) Error.CodeError
}

func (m *mockPublisher) SendVerificationEmail(ctx context.Context, dto entities.VerificationEmailMsg) Error.CodeError {
	return m.sendVerificationEmail(ctx, dto)
}
func (m *mockPublisher) SendRecoveryEmail(ctx context.Context, dto entities.RecoveryEmailMsg) Error.CodeError {
	return m.sendRecoveryEmail(ctx, dto)
}

// emptyPublisher — заглушка для тестов, где Publisher не должен вызываться.
func emptyPublisher() messaging.Publisher {
	return &mockPublisher{
		sendVerificationEmail: func(_ context.Context, _ entities.VerificationEmailMsg) Error.CodeError { return Error.CodeError{} },
		sendRecoveryEmail:     func(_ context.Context, _ entities.RecoveryEmailMsg) Error.CodeError { return Error.CodeError{} },
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// testPrivateKey — ECDSA ключ из /keys/test/private.pem для unit-тестов.
var testPrivateKey *ecdsa.PrivateKey

func init() {
	var err error
	testPrivateKey, err = utils.LoadPrivateKey("../../../keys/test/private.pem")
	if err != nil {
		panic("failed to load test private key: " + err.Error())
	}
}

const (
	testAccessTTL  = time.Minute
	testRefreshTTL = time.Hour

	// Валидные UUID для использования в тестах
	testUUID1 = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	testUUID2 = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	// Валидный пароль: есть uppercase, lowercase и цифра, длина >= 8
	testPassword = "Password123"
)

// newTestService создаёт AuthService с подменёнными зависимостями
func newTestService(userRepo postgresDB.UserRepository, authRepo redisDB.AuthRepository) *AuthService {
	db := &postgresDB.DatabaseRepository{User: userRepo}
	cache := &redisDB.CacheRepository{Auth: authRepo, Verification: emptyVerificationRepo()}
	return NewAuthService(db, cache, emptyPublisher(), testPrivateKey, testAccessTTL, testRefreshTTL, "test")
}

// emptyUserRepo — заглушка для тестов, где UserRepository не должен вызываться
func emptyUserRepo() *mockUserRepo { return &mockUserRepo{} }

// emptyAuthRepo — заглушка для тестов, где AuthRepository не должен вызываться
func emptyAuthRepo() *mockAuthRepo { return &mockAuthRepo{} }

// emptyVerificationRepo — заглушка для тестов, где VerificationRepository не должен вызываться.
// Все операции возвращают успех, чтобы не мешать тестам Register.
func emptyVerificationRepo() *mockVerificationRepo {
	return &mockVerificationRepo{
		saveVerificationCode:     func(_ context.Context, _ entities.SaveVerificationCodeDTO) Error.CodeError { return Error.CodeError{} },
		getVerificationCode:      func(_ context.Context, _ entities.GetVerificationCodeDTO) (string, Error.CodeError) { return "", Error.CodeError{} },
		deleteVerificationCode:   func(_ context.Context, _ entities.DeleteVerificationCodeDTO) Error.CodeError { return Error.CodeError{} },
		incrVerificationAttempts: func(_ context.Context, _ entities.IncrVerificationAttemptsDTO) (int64, Error.CodeError) { return 0, Error.CodeError{} },
		saveRecoveryCode:         func(_ context.Context, _ entities.SaveRecoveryCodeDTO) Error.CodeError { return Error.CodeError{} },
		getRecoveryCode:          func(_ context.Context, _ entities.GetRecoveryCodeDTO) (string, Error.CodeError) { return "", Error.CodeError{} },
		deleteRecoveryCode:       func(_ context.Context, _ entities.DeleteRecoveryCodeDTO) Error.CodeError { return Error.CodeError{} },
		incrRecoveryAttempts:     func(_ context.Context, _ entities.IncrRecoveryAttemptsDTO) (int64, Error.CodeError) { return 0, Error.CodeError{} },
	}
}

// ok возвращает CodeError без ошибки
func ok() Error.CodeError { return Error.CodeError{} }
