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
	createUser           func(ctx context.Context, dto entities.User) Error.CodeError
	getUserByEmail       func(ctx context.Context, dto entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError)
	getUser              func(ctx context.Context, dto entities.GetUserDTO) (*entities.UserGet, Error.CodeError)
	updateUserPassword   func(ctx context.Context, dto entities.UpdateUserPasswordDTO) Error.CodeError
	updateUserBio        func(ctx context.Context, dto entities.UserUpdateBioDTO) Error.CodeError
	updateUser2FA        func(ctx context.Context, dto entities.UpdateUser2FADTO) Error.CodeError
	deleteUser           func(ctx context.Context, dto entities.DeleteUserDTO) Error.CodeError
	restoreUser          func(ctx context.Context, dto entities.RestoreUserDTO) Error.CodeError
	anonymizeExpiredUsers func(ctx context.Context, before time.Time) (int64, error)
	setUserVerified      func(ctx context.Context, dto entities.SetUserVerifiedDTO) Error.CodeError
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
func (m *mockUserRepo) UpdateUserBio(ctx context.Context, dto entities.UserUpdateBioDTO) Error.CodeError {
	return m.updateUserBio(ctx, dto)
}
func (m *mockUserRepo) DeleteUser(ctx context.Context, dto entities.DeleteUserDTO) Error.CodeError {
	return m.deleteUser(ctx, dto)
}
func (m *mockUserRepo) RestoreUser(ctx context.Context, dto entities.RestoreUserDTO) Error.CodeError {
	return m.restoreUser(ctx, dto)
}
func (m *mockUserRepo) AnonymizeExpiredUsers(ctx context.Context, before time.Time) (int64, error) {
	return m.anonymizeExpiredUsers(ctx, before)
}
func (m *mockUserRepo) UpdateUser2FA(ctx context.Context, dto entities.UpdateUser2FADTO) Error.CodeError {
	return m.updateUser2FA(ctx, dto)
}
func (m *mockUserRepo) SetUserVerified(ctx context.Context, dto entities.SetUserVerifiedDTO) Error.CodeError {
	return m.setUserVerified(ctx, dto)
}

// ─── Mock: AuthRepository ────────────────────────────────────────────────────

type mockAuthRepo struct {
	saveSession        func(ctx context.Context, dto entities.SaveSessionDTO) Error.CodeError
	getAllSessions      func(ctx context.Context, dto entities.GetAllSessionsDTO) ([]entities.SessionEntry, Error.CodeError)
	checkSessionExists func(ctx context.Context, dto entities.CheckSessionExistsDTO) Error.CodeError
	revokeSession      func(ctx context.Context, dto entities.RevokeSessionDTO) Error.CodeError
	revokeAllSessions  func(ctx context.Context, dto entities.RevokeAllSessionsDTO) Error.CodeError
	refreshToken      func(ctx context.Context, dto entities.RefreshTokenDTO) Error.CodeError
}

func (m *mockAuthRepo) SaveSession(ctx context.Context, dto entities.SaveSessionDTO) Error.CodeError {
	return m.saveSession(ctx, dto)
}
func (m *mockAuthRepo) GetAllSessions(ctx context.Context, dto entities.GetAllSessionsDTO) ([]entities.SessionEntry, Error.CodeError) {
	return m.getAllSessions(ctx, dto)
}
func (m *mockAuthRepo) CheckSessionExists(ctx context.Context, dto entities.CheckSessionExistsDTO) Error.CodeError {
	return m.checkSessionExists(ctx, dto)
}
func (m *mockAuthRepo) RevokeSession(ctx context.Context, dto entities.RevokeSessionDTO) Error.CodeError {
	return m.revokeSession(ctx, dto)
}
func (m *mockAuthRepo) RevokeAllSessions(ctx context.Context, dto entities.RevokeAllSessionsDTO) Error.CodeError {
	return m.revokeAllSessions(ctx, dto)
}
func (m *mockAuthRepo) RefreshToken(ctx context.Context, dto entities.RefreshTokenDTO) Error.CodeError {
	return m.refreshToken(ctx, dto)
}

// ─── Mock: VerificationRepository ───────────────────────────────────────────

type mockVerificationRepo struct {
	addToVerificationTokenBlacklist  func(ctx context.Context, dto entities.AddToVerificationTokenBlacklistDTO) Error.CodeError
	isVerificationTokenBlacklisted   func(ctx context.Context, dto entities.IsVerificationTokenBlacklistedDTO) (bool, Error.CodeError)
	acquireVerificationEmailCooldown func(ctx context.Context, dto entities.AcquireVerificationEmailCooldownDTO) (bool, Error.CodeError)
	incrVerificationEmailDailyCount  func(ctx context.Context, dto entities.IncrVerificationEmailDailyCountDTO) (int64, Error.CodeError)
}

func (m *mockVerificationRepo) AddToVerificationTokenBlacklist(ctx context.Context, dto entities.AddToVerificationTokenBlacklistDTO) Error.CodeError {
	return m.addToVerificationTokenBlacklist(ctx, dto)
}
func (m *mockVerificationRepo) IsVerificationTokenBlacklisted(ctx context.Context, dto entities.IsVerificationTokenBlacklistedDTO) (bool, Error.CodeError) {
	return m.isVerificationTokenBlacklisted(ctx, dto)
}
func (m *mockVerificationRepo) AcquireVerificationEmailCooldown(ctx context.Context, dto entities.AcquireVerificationEmailCooldownDTO) (bool, Error.CodeError) {
	return m.acquireVerificationEmailCooldown(ctx, dto)
}
func (m *mockVerificationRepo) IncrVerificationEmailDailyCount(ctx context.Context, dto entities.IncrVerificationEmailDailyCountDTO) (int64, Error.CodeError) {
	return m.incrVerificationEmailDailyCount(ctx, dto)
}

// ─── Mock: RecoveryRepository ────────────────────────────────────────────────

type mockRecoveryRepo struct {
	addToResetTokenBlacklist     func(ctx context.Context, dto entities.AddToResetTokenBlacklistDTO) Error.CodeError
	isResetTokenBlacklisted      func(ctx context.Context, dto entities.IsResetTokenBlacklistedDTO) (bool, Error.CodeError)
	acquireRecoveryEmailCooldown func(ctx context.Context, dto entities.AcquireRecoveryEmailCooldownDTO) (bool, Error.CodeError)
	incrRecoveryEmailDailyCount  func(ctx context.Context, dto entities.IncrRecoveryEmailDailyCountDTO) (int64, Error.CodeError)
}

func (m *mockRecoveryRepo) AddToResetTokenBlacklist(ctx context.Context, dto entities.AddToResetTokenBlacklistDTO) Error.CodeError {
	return m.addToResetTokenBlacklist(ctx, dto)
}
func (m *mockRecoveryRepo) IsResetTokenBlacklisted(ctx context.Context, dto entities.IsResetTokenBlacklistedDTO) (bool, Error.CodeError) {
	return m.isResetTokenBlacklisted(ctx, dto)
}
func (m *mockRecoveryRepo) AcquireRecoveryEmailCooldown(ctx context.Context, dto entities.AcquireRecoveryEmailCooldownDTO) (bool, Error.CodeError) {
	return m.acquireRecoveryEmailCooldown(ctx, dto)
}
func (m *mockRecoveryRepo) IncrRecoveryEmailDailyCount(ctx context.Context, dto entities.IncrRecoveryEmailDailyCountDTO) (int64, Error.CodeError) {
	return m.incrRecoveryEmailDailyCount(ctx, dto)
}

// emptyRecoveryRepo — заглушка для тестов, где RecoveryRepository не должен вызываться.
func emptyRecoveryRepo() *mockRecoveryRepo {
	return &mockRecoveryRepo{
		addToResetTokenBlacklist: func(_ context.Context, _ entities.AddToResetTokenBlacklistDTO) Error.CodeError {
			return Error.CodeError{}
		},
		isResetTokenBlacklisted: func(_ context.Context, _ entities.IsResetTokenBlacklistedDTO) (bool, Error.CodeError) {
			return false, Error.CodeError{}
		},
		acquireRecoveryEmailCooldown: func(_ context.Context, _ entities.AcquireRecoveryEmailCooldownDTO) (bool, Error.CodeError) {
			return true, Error.CodeError{}
		},
		incrRecoveryEmailDailyCount: func(_ context.Context, _ entities.IncrRecoveryEmailDailyCountDTO) (int64, Error.CodeError) {
			return 1, Error.CodeError{}
		},
	}
}

// ─── Mock: TwoFARepository ───────────────────────────────────────────────────

type mockTwoFARepo struct {
	save2FAData            func(ctx context.Context, dto entities.Save2FADataDTO) Error.CodeError
	get2FAData             func(ctx context.Context, dto entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError)
	delete2FAData          func(ctx context.Context, dto entities.Delete2FADataDTO) Error.CodeError
	incr2FAAttempts        func(ctx context.Context, dto entities.Incr2FAAttemptsDTO) (int64, Error.CodeError)
	acquire2FAEmailCooldown func(ctx context.Context, dto entities.Acquire2FAEmailCooldownDTO) (bool, Error.CodeError)
	incr2FAEmailDailyCount  func(ctx context.Context, dto entities.Incr2FAEmailDailyCountDTO) (int64, Error.CodeError)
}

func (m *mockTwoFARepo) Save2FAData(ctx context.Context, dto entities.Save2FADataDTO) Error.CodeError {
	return m.save2FAData(ctx, dto)
}
func (m *mockTwoFARepo) Get2FAData(ctx context.Context, dto entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError) {
	return m.get2FAData(ctx, dto)
}
func (m *mockTwoFARepo) Delete2FAData(ctx context.Context, dto entities.Delete2FADataDTO) Error.CodeError {
	return m.delete2FAData(ctx, dto)
}
func (m *mockTwoFARepo) Incr2FAAttempts(ctx context.Context, dto entities.Incr2FAAttemptsDTO) (int64, Error.CodeError) {
	return m.incr2FAAttempts(ctx, dto)
}
func (m *mockTwoFARepo) Acquire2FAEmailCooldown(ctx context.Context, dto entities.Acquire2FAEmailCooldownDTO) (bool, Error.CodeError) {
	return m.acquire2FAEmailCooldown(ctx, dto)
}
func (m *mockTwoFARepo) Incr2FAEmailDailyCount(ctx context.Context, dto entities.Incr2FAEmailDailyCountDTO) (int64, Error.CodeError) {
	return m.incr2FAEmailDailyCount(ctx, dto)
}

// emptyTwoFARepo — заглушка для тестов, где TwoFARepository не должен вызываться.
func emptyTwoFARepo() *mockTwoFARepo {
	return &mockTwoFARepo{
		save2FAData:             func(_ context.Context, _ entities.Save2FADataDTO) Error.CodeError { return Error.CodeError{} },
		get2FAData:              func(_ context.Context, _ entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError) { return nil, Error.CodeError{} },
		delete2FAData:           func(_ context.Context, _ entities.Delete2FADataDTO) Error.CodeError { return Error.CodeError{} },
		incr2FAAttempts:         func(_ context.Context, _ entities.Incr2FAAttemptsDTO) (int64, Error.CodeError) { return 0, Error.CodeError{} },
		acquire2FAEmailCooldown: func(_ context.Context, _ entities.Acquire2FAEmailCooldownDTO) (bool, Error.CodeError) { return true, Error.CodeError{} },
		incr2FAEmailDailyCount:  func(_ context.Context, _ entities.Incr2FAEmailDailyCountDTO) (int64, Error.CodeError) { return 1, Error.CodeError{} },
	}
}

// ─── Mock: Publisher ─────────────────────────────────────────────────────────

type mockPublisher struct {
	sendVerificationEmail        func(ctx context.Context, dto entities.VerificationEmailMsg) Error.CodeError
	sendRecoveryEmail            func(ctx context.Context, dto entities.RecoveryEmailMsg) Error.CodeError
	send2FAEmail                 func(ctx context.Context, dto entities.TwoFAEmailMsg) Error.CodeError
	sendPasswordChangedEmail     func(ctx context.Context, dto entities.PasswordChangedEmailMsg) Error.CodeError
	sendPasswordResetEmail       func(ctx context.Context, dto entities.PasswordResetEmailMsg) Error.CodeError
	sendRegistrationAttemptEmail func(ctx context.Context, dto entities.RegistrationAttemptEmailMsg) Error.CodeError
	sendLoginNotificationEmail   func(ctx context.Context, dto entities.LoginNotificationEmailMsg) Error.CodeError
	sendTokenReuseAlertEmail     func(ctx context.Context, dto entities.TokenReuseAlertEmailMsg) Error.CodeError
}

func (m *mockPublisher) SendVerificationEmail(ctx context.Context, dto entities.VerificationEmailMsg) Error.CodeError {
	return m.sendVerificationEmail(ctx, dto)
}
func (m *mockPublisher) SendRecoveryEmail(ctx context.Context, dto entities.RecoveryEmailMsg) Error.CodeError {
	return m.sendRecoveryEmail(ctx, dto)
}
func (m *mockPublisher) Send2FAEmail(ctx context.Context, dto entities.TwoFAEmailMsg) Error.CodeError {
	return m.send2FAEmail(ctx, dto)
}
func (m *mockPublisher) SendPasswordChangedEmail(ctx context.Context, dto entities.PasswordChangedEmailMsg) Error.CodeError {
	return m.sendPasswordChangedEmail(ctx, dto)
}
func (m *mockPublisher) SendPasswordResetEmail(ctx context.Context, dto entities.PasswordResetEmailMsg) Error.CodeError {
	return m.sendPasswordResetEmail(ctx, dto)
}
func (m *mockPublisher) SendRegistrationAttemptEmail(ctx context.Context, dto entities.RegistrationAttemptEmailMsg) Error.CodeError {
	return m.sendRegistrationAttemptEmail(ctx, dto)
}
func (m *mockPublisher) SendLoginNotificationEmail(ctx context.Context, dto entities.LoginNotificationEmailMsg) Error.CodeError {
	return m.sendLoginNotificationEmail(ctx, dto)
}
func (m *mockPublisher) SendTokenReuseAlertEmail(ctx context.Context, dto entities.TokenReuseAlertEmailMsg) Error.CodeError {
	return m.sendTokenReuseAlertEmail(ctx, dto)
}

// emptyPublisher — заглушка для тестов, где Publisher не должен вызываться.
func emptyPublisher() messaging.Publisher {
	return &mockPublisher{
		sendVerificationEmail:        func(_ context.Context, _ entities.VerificationEmailMsg) Error.CodeError { return Error.CodeError{} },
		sendRecoveryEmail:            func(_ context.Context, _ entities.RecoveryEmailMsg) Error.CodeError { return Error.CodeError{} },
		send2FAEmail:                 func(_ context.Context, _ entities.TwoFAEmailMsg) Error.CodeError { return Error.CodeError{} },
		sendPasswordChangedEmail:     func(_ context.Context, _ entities.PasswordChangedEmailMsg) Error.CodeError { return Error.CodeError{} },
		sendPasswordResetEmail:       func(_ context.Context, _ entities.PasswordResetEmailMsg) Error.CodeError { return Error.CodeError{} },
		sendRegistrationAttemptEmail: func(_ context.Context, _ entities.RegistrationAttemptEmailMsg) Error.CodeError { return Error.CodeError{} },
		sendLoginNotificationEmail:   func(_ context.Context, _ entities.LoginNotificationEmailMsg) Error.CodeError { return Error.CodeError{} },
		sendTokenReuseAlertEmail:     func(_ context.Context, _ entities.TokenReuseAlertEmailMsg) Error.CodeError { return Error.CodeError{} },
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
	cache := &redisDB.CacheRepository{
		Auth:         authRepo,
		Verification: emptyVerificationRepo(),
		Recovery:     emptyRecoveryRepo(),
		TwoFA:        emptyTwoFARepo(),
	}
	return NewAuthService(db, cache, emptyPublisher(), testPrivateKey, testAccessTTL, testRefreshTTL, "test")
}

// emptyUserRepo — заглушка для тестов, где UserRepository не должен вызываться
func emptyUserRepo() *mockUserRepo { return &mockUserRepo{} }

// emptyAuthRepo — заглушка для тестов, где AuthRepository не должен вызываться
func emptyAuthRepo() *mockAuthRepo { return &mockAuthRepo{} }

// emptyVerificationRepo — заглушка для тестов, где VerificationRepository не должен вызываться.
func emptyVerificationRepo() *mockVerificationRepo {
	return &mockVerificationRepo{
		addToVerificationTokenBlacklist: func(_ context.Context, _ entities.AddToVerificationTokenBlacklistDTO) Error.CodeError {
			return Error.CodeError{}
		},
		isVerificationTokenBlacklisted: func(_ context.Context, _ entities.IsVerificationTokenBlacklistedDTO) (bool, Error.CodeError) {
			return false, Error.CodeError{}
		},
		acquireVerificationEmailCooldown: func(_ context.Context, _ entities.AcquireVerificationEmailCooldownDTO) (bool, Error.CodeError) {
			return true, Error.CodeError{}
		},
		incrVerificationEmailDailyCount: func(_ context.Context, _ entities.IncrVerificationEmailDailyCountDTO) (int64, Error.CodeError) {
			return 1, Error.CodeError{}
		},
	}
}

// ok возвращает CodeError без ошибки
func ok() Error.CodeError { return Error.CodeError{} }

// ─── Flexible service builder ─────────────────────────────────────────────────

// svcDeps хранит зависимости для сборки тестового AuthService.
// Любое nil-поле заменяется пустой заглушкой.
type svcDeps struct {
	user         postgresDB.UserRepository
	auth         redisDB.AuthRepository
	verification redisDB.VerificationRepository
	recovery     redisDB.RecoveryRepository
	twoFA        redisDB.TwoFARepository
	publisher    messaging.Publisher
	appEnv       string
}

// buildSvc создаёт AuthService с заданными зависимостями.
func buildSvc(d svcDeps) *AuthService {
	if d.user == nil {
		d.user = emptyUserRepo()
	}
	if d.auth == nil {
		d.auth = emptyAuthRepo()
	}
	if d.verification == nil {
		d.verification = emptyVerificationRepo()
	}
	if d.recovery == nil {
		d.recovery = emptyRecoveryRepo()
	}
	if d.twoFA == nil {
		d.twoFA = emptyTwoFARepo()
	}
	if d.publisher == nil {
		d.publisher = emptyPublisher()
	}
	if d.appEnv == "" {
		d.appEnv = "test"
	}
	db := &postgresDB.DatabaseRepository{User: d.user}
	cache := &redisDB.CacheRepository{
		Auth:         d.auth,
		Verification: d.verification,
		Recovery:     d.recovery,
		TwoFA:        d.twoFA,
	}
	return NewAuthService(db, cache, d.publisher, testPrivateKey, testAccessTTL, testRefreshTTL, d.appEnv)
}
