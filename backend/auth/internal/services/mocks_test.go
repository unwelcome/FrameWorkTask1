package services

import (
	"context"
	"time"

	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
)

// ─── Mock: UserRepository ────────────────────────────────────────────────────

type mockUserRepo struct {
	createUser         func(ctx context.Context, dto *entities.UserCreate) Error.CodeError
	getUserByEmail     func(ctx context.Context, email string) (*entities.UserGetByEmail, Error.CodeError)
	getUser            func(ctx context.Context, uuid string) (*entities.UserGet, Error.CodeError)
	updateUserPassword func(ctx context.Context, uuid, passwordHash string) Error.CodeError
	updateUserBio      func(ctx context.Context, dto *entities.UserUpdateBio) Error.CodeError
	deleteUser         func(ctx context.Context, uuid string) Error.CodeError
}

func (m *mockUserRepo) CreateUser(ctx context.Context, dto *entities.UserCreate) Error.CodeError {
	return m.createUser(ctx, dto)
}
func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*entities.UserGetByEmail, Error.CodeError) {
	return m.getUserByEmail(ctx, email)
}
func (m *mockUserRepo) GetUser(ctx context.Context, uuid string) (*entities.UserGet, Error.CodeError) {
	return m.getUser(ctx, uuid)
}
func (m *mockUserRepo) UpdateUserPassword(ctx context.Context, uuid, passwordHash string) Error.CodeError {
	return m.updateUserPassword(ctx, uuid, passwordHash)
}
func (m *mockUserRepo) UpdateUserBio(ctx context.Context, dto *entities.UserUpdateBio) Error.CodeError {
	return m.updateUserBio(ctx, dto)
}
func (m *mockUserRepo) DeleteUser(ctx context.Context, uuid string) Error.CodeError {
	return m.deleteUser(ctx, uuid)
}

// ─── Mock: AuthRepository ────────────────────────────────────────────────────

type mockAuthRepo struct {
	saveRefreshToken        func(ctx context.Context, userUUID, rawToken string) Error.CodeError
	getAllRefreshTokens     func(ctx context.Context, userUUID string) ([]string, Error.CodeError)
	checkRefreshTokenExists func(ctx context.Context, userUUID, rawToken string) Error.CodeError
	revokeRefreshToken      func(ctx context.Context, userUUID, tokenHash string) Error.CodeError
	revokeAllRefreshTokens  func(ctx context.Context, userUUID string) Error.CodeError
	refreshToken            func(ctx context.Context, userUUID, oldRawToken, newRawToken string) Error.CodeError
}

func (m *mockAuthRepo) SaveRefreshToken(ctx context.Context, userUUID, rawToken string) Error.CodeError {
	return m.saveRefreshToken(ctx, userUUID, rawToken)
}
func (m *mockAuthRepo) GetAllRefreshTokens(ctx context.Context, userUUID string) ([]string, Error.CodeError) {
	return m.getAllRefreshTokens(ctx, userUUID)
}
func (m *mockAuthRepo) CheckRefreshTokenExists(ctx context.Context, userUUID, rawToken string) Error.CodeError {
	return m.checkRefreshTokenExists(ctx, userUUID, rawToken)
}
func (m *mockAuthRepo) RevokeRefreshToken(ctx context.Context, userUUID, tokenHash string) Error.CodeError {
	return m.revokeRefreshToken(ctx, userUUID, tokenHash)
}
func (m *mockAuthRepo) RevokeAllRefreshTokens(ctx context.Context, userUUID string) Error.CodeError {
	return m.revokeAllRefreshTokens(ctx, userUUID)
}
func (m *mockAuthRepo) RefreshToken(ctx context.Context, userUUID, oldRawToken, newRawToken string) Error.CodeError {
	return m.refreshToken(ctx, userUUID, oldRawToken, newRawToken)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

const (
	testSecret     = "test-secret-key"
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
	return NewAuthService(db, cache, testSecret, testAccessTTL, testRefreshTTL)
}

// emptyUserRepo — заглушка для тестов, где UserRepository не должен вызываться
func emptyUserRepo() *mockUserRepo { return &mockUserRepo{} }

// emptyAuthRepo — заглушка для тестов, где AuthRepository не должен вызываться
func emptyAuthRepo() *mockAuthRepo { return &mockAuthRepo{} }

// ok возвращает CodeError без ошибки
func ok() Error.CodeError { return Error.CodeError{} }
