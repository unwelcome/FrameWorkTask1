package services

import (
	"context"
	"crypto/ecdsa"
	"time"

	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/utils"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
)

// ─── Mock: UserRepository ────────────────────────────────────────────────────

type mockUserRepo struct {
	createUser         func(ctx context.Context, dto *entities.User) Error.CodeError
	getUserByEmail     func(ctx context.Context, dto entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError)
	getUser            func(ctx context.Context, dto entities.GetUserDTO) (*entities.UserGet, Error.CodeError)
	updateUserPassword func(ctx context.Context, dto entities.UpdateUserPasswordDTO) Error.CodeError
	updateUserBio      func(ctx context.Context, dto *entities.UserUpdateBio) Error.CodeError
	deleteUser         func(ctx context.Context, dto entities.DeleteUserDTO) Error.CodeError
}

func (m *mockUserRepo) CreateUser(ctx context.Context, dto *entities.User) Error.CodeError {
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
func (m *mockUserRepo) UpdateUserBio(ctx context.Context, dto *entities.UserUpdateBio) Error.CodeError {
	return m.updateUserBio(ctx, dto)
}
func (m *mockUserRepo) DeleteUser(ctx context.Context, dto entities.DeleteUserDTO) Error.CodeError {
	return m.deleteUser(ctx, dto)
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
	cache := &redisDB.CacheRepository{Auth: authRepo}
	return NewAuthService(db, cache, testPrivateKey, testAccessTTL, testRefreshTTL)
}

// emptyUserRepo — заглушка для тестов, где UserRepository не должен вызываться
func emptyUserRepo() *mockUserRepo { return &mockUserRepo{} }

// emptyAuthRepo — заглушка для тестов, где AuthRepository не должен вызываться
func emptyAuthRepo() *mockAuthRepo { return &mockAuthRepo{} }

// ok возвращает CodeError без ошибки
func ok() Error.CodeError { return Error.CodeError{} }
