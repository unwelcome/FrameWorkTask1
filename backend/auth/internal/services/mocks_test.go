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
	saveRefreshToken        func(ctx context.Context, dto entities.SaveRefreshTokenDTO) Error.CodeError
	getAllRefreshTokens     func(ctx context.Context, dto entities.GetAllRefreshTokensDTO) ([]entities.RefreshTokenEntry, Error.CodeError)
	checkRefreshTokenExists func(ctx context.Context, dto entities.CheckRefreshTokenExistsDTO) Error.CodeError
	revokeRefreshToken      func(ctx context.Context, dto entities.RevokeRefreshTokenDTO) Error.CodeError
	revokeAllRefreshTokens  func(ctx context.Context, dto entities.RevokeAllRefreshTokensDTO) Error.CodeError
	refreshToken            func(ctx context.Context, dto entities.RefreshTokenDTO) Error.CodeError
}

func (m *mockAuthRepo) SaveRefreshToken(ctx context.Context, dto entities.SaveRefreshTokenDTO) Error.CodeError {
	return m.saveRefreshToken(ctx, dto)
}
func (m *mockAuthRepo) GetAllRefreshTokens(ctx context.Context, dto entities.GetAllRefreshTokensDTO) ([]entities.RefreshTokenEntry, Error.CodeError) {
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
	acquireResendCooldown    func(ctx context.Context, dto entities.CheckResendCooldownDTO) (bool, Error.CodeError)
	incrResendDailyCount     func(ctx context.Context, dto entities.IncrResendDailyCountDTO) (int64, Error.CodeError)

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
func (m *mockVerificationRepo) AcquireResendCooldown(ctx context.Context, dto entities.CheckResendCooldownDTO) (bool, Error.CodeError) {
	return m.acquireResendCooldown(ctx, dto)
}
func (m *mockVerificationRepo) IncrResendDailyCount(ctx context.Context, dto entities.IncrResendDailyCountDTO) (int64, Error.CodeError) {
	return m.incrResendDailyCount(ctx, dto)
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

// ─── Mock: RecoveryRepository ────────────────────────────────────────────────

type mockRecoveryRepo struct {
	saveRecoveryCode     func(ctx context.Context, dto entities.SaveRecoveryCodeDTO) Error.CodeError
	getRecoveryCode      func(ctx context.Context, dto entities.GetRecoveryCodeDTO) (string, Error.CodeError)
	deleteRecoveryCode   func(ctx context.Context, dto entities.DeleteRecoveryCodeDTO) Error.CodeError
	incrRecoveryAttempts func(ctx context.Context, dto entities.IncrRecoveryAttemptsDTO) (int64, Error.CodeError)
}

func (m *mockRecoveryRepo) SaveRecoveryCode(ctx context.Context, dto entities.SaveRecoveryCodeDTO) Error.CodeError {
	return m.saveRecoveryCode(ctx, dto)
}
func (m *mockRecoveryRepo) GetRecoveryCode(ctx context.Context, dto entities.GetRecoveryCodeDTO) (string, Error.CodeError) {
	return m.getRecoveryCode(ctx, dto)
}
func (m *mockRecoveryRepo) DeleteRecoveryCode(ctx context.Context, dto entities.DeleteRecoveryCodeDTO) Error.CodeError {
	return m.deleteRecoveryCode(ctx, dto)
}
func (m *mockRecoveryRepo) IncrRecoveryAttempts(ctx context.Context, dto entities.IncrRecoveryAttemptsDTO) (int64, Error.CodeError) {
	return m.incrRecoveryAttempts(ctx, dto)
}

// emptyRecoveryRepo — заглушка для тестов, где RecoveryRepository не должен вызываться.
func emptyRecoveryRepo() *mockRecoveryRepo {
	return &mockRecoveryRepo{
		saveRecoveryCode:     func(_ context.Context, _ entities.SaveRecoveryCodeDTO) Error.CodeError { return Error.CodeError{} },
		getRecoveryCode:      func(_ context.Context, _ entities.GetRecoveryCodeDTO) (string, Error.CodeError) { return "", Error.CodeError{} },
		deleteRecoveryCode:   func(_ context.Context, _ entities.DeleteRecoveryCodeDTO) Error.CodeError { return Error.CodeError{} },
		incrRecoveryAttempts: func(_ context.Context, _ entities.IncrRecoveryAttemptsDTO) (int64, Error.CodeError) { return 0, Error.CodeError{} },
	}
}

// ─── Mock: TwoFARepository ───────────────────────────────────────────────────

type mockTwoFARepo struct {
	save2FAData     func(ctx context.Context, dto entities.Save2FADataDTO) Error.CodeError
	get2FAData      func(ctx context.Context, dto entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError)
	delete2FAData   func(ctx context.Context, dto entities.Delete2FADataDTO) Error.CodeError
	incr2FAAttempts func(ctx context.Context, dto entities.Incr2FAAttemptsDTO) (int64, Error.CodeError)
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

// emptyTwoFARepo — заглушка для тестов, где TwoFARepository не должен вызываться.
func emptyTwoFARepo() *mockTwoFARepo {
	return &mockTwoFARepo{
		save2FAData:     func(_ context.Context, _ entities.Save2FADataDTO) Error.CodeError { return Error.CodeError{} },
		get2FAData:      func(_ context.Context, _ entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError) { return nil, Error.CodeError{} },
		delete2FAData:   func(_ context.Context, _ entities.Delete2FADataDTO) Error.CodeError { return Error.CodeError{} },
		incr2FAAttempts: func(_ context.Context, _ entities.Incr2FAAttemptsDTO) (int64, Error.CodeError) { return 0, Error.CodeError{} },
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

// emptyPublisher — заглушка для тестов, где Publisher не должен вызываться.
func emptyPublisher() messaging.Publisher {
	return &mockPublisher{
		sendVerificationEmail:        func(_ context.Context, _ entities.VerificationEmailMsg) Error.CodeError { return Error.CodeError{} },
		sendRecoveryEmail:            func(_ context.Context, _ entities.RecoveryEmailMsg) Error.CodeError { return Error.CodeError{} },
		send2FAEmail:                 func(_ context.Context, _ entities.TwoFAEmailMsg) Error.CodeError { return Error.CodeError{} },
		sendPasswordChangedEmail:     func(_ context.Context, _ entities.PasswordChangedEmailMsg) Error.CodeError { return Error.CodeError{} },
		sendPasswordResetEmail:       func(_ context.Context, _ entities.PasswordResetEmailMsg) Error.CodeError { return Error.CodeError{} },
		sendRegistrationAttemptEmail: func(_ context.Context, _ entities.RegistrationAttemptEmailMsg) Error.CodeError { return Error.CodeError{} },
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
// Все операции возвращают успех, чтобы не мешать тестам Register.
func emptyVerificationRepo() *mockVerificationRepo {
	return &mockVerificationRepo{
		saveVerificationCode: func(_ context.Context, _ entities.SaveVerificationCodeDTO) Error.CodeError { return Error.CodeError{} },
		getVerificationCode: func(_ context.Context, _ entities.GetVerificationCodeDTO) (string, Error.CodeError) {
			return "", Error.CodeError{}
		},
		deleteVerificationCode: func(_ context.Context, _ entities.DeleteVerificationCodeDTO) Error.CodeError {
			return Error.CodeError{}
		},
		incrVerificationAttempts: func(_ context.Context, _ entities.IncrVerificationAttemptsDTO) (int64, Error.CodeError) {
			return 0, Error.CodeError{}
		},
		acquireResendCooldown: func(_ context.Context, _ entities.CheckResendCooldownDTO) (bool, Error.CodeError) {
			return true, Error.CodeError{}
		},
		incrResendDailyCount: func(_ context.Context, _ entities.IncrResendDailyCountDTO) (int64, Error.CodeError) {
			return 1, Error.CodeError{}
		},
		saveRecoveryCode: func(_ context.Context, _ entities.SaveRecoveryCodeDTO) Error.CodeError { return Error.CodeError{} },
		getRecoveryCode: func(_ context.Context, _ entities.GetRecoveryCodeDTO) (string, Error.CodeError) {
			return "", Error.CodeError{}
		},
		deleteRecoveryCode: func(_ context.Context, _ entities.DeleteRecoveryCodeDTO) Error.CodeError { return Error.CodeError{} },
		incrRecoveryAttempts: func(_ context.Context, _ entities.IncrRecoveryAttemptsDTO) (int64, Error.CodeError) {
			return 0, Error.CodeError{}
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
