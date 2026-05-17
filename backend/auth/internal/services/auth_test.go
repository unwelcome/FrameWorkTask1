package services

import (
	"context"
	"fmt"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/unwelcome/FrameWorkTask1/backend/auth/api/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/utils"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
)

// assertCode проверяет, что err — gRPC-ошибка с ожидаемым кодом
func assertCode(t *testing.T, err error, expected codes.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected gRPC error %v, got nil", expected)
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status error: %v", err)
	}
	if st.Code() != expected {
		t.Fatalf("expected gRPC code %v, got %v (msg: %q)", expected, st.Code(), st.Message())
	}
}

// assertNoError проверяет отсутствие ошибки
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// hashPassword хеширует пароль для использования в тестах
func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hash)
}

// validRefreshToken генерирует настоящий refresh-токен для тестов
func validRefreshToken(t *testing.T) string {
	t.Helper()
	tokens, err := utils.CreateTokens("user-uuid-1", testSecret, testAccessTTL, testRefreshTTL)
	if err != nil {
		t.Fatalf("failed to create tokens: %v", err)
	}
	return tokens.RefreshToken
}

// validAccessToken генерирует настоящий access-токен для тестов
func validAccessToken(t *testing.T) string {
	t.Helper()
	tokens, err := utils.CreateTokens("user-uuid-1", testSecret, testAccessTTL, testRefreshTTL)
	if err != nil {
		t.Fatalf("failed to create tokens: %v", err)
	}
	return tokens.AccessToken
}

// ─── Register ────────────────────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	userRepo := &mockUserRepo{
		createUser: func(_ context.Context, dto *entities.UserCreate) Error.CodeError {
			if dto.Email != "test@example.com" {
				return Error.CodeError{Code: 0, Err: fmt.Errorf("unexpected email")}
			}
			return ok()
		},
	}
	svc := newTestService(userRepo, emptyAuthRepo())

	resp, err := svc.Register(context.Background(), &pb.RegisterRequest{
		OperationId: "op-1",
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "Ivan",
		LastName:    "Ivanov",
	})

	assertNoError(t, err)
	if resp.GetUserUuid() == "" {
		t.Error("expected non-empty user UUID")
	}
}

func TestRegister_PasswordTooLong(t *testing.T) {
	svc := newTestService(emptyUserRepo(), emptyAuthRepo())

	_, err := svc.Register(context.Background(), &pb.RegisterRequest{
		OperationId: "op-1",
		Password:    string(make([]byte, 70)),
	})

	assertCode(t, err, codes.InvalidArgument)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	userRepo := &mockUserRepo{
		createUser: func(_ context.Context, _ *entities.UserCreate) Error.CodeError {
			return Error.CodeError{Code: int(codes.AlreadyExists), Err: fmt.Errorf("email already registered")}
		},
	}
	svc := newTestService(userRepo, emptyAuthRepo())

	_, err := svc.Register(context.Background(), &pb.RegisterRequest{
		OperationId: "op-1",
		Email:       "taken@example.com",
		Password:    "password123",
		FirstName:   "Ivan",
		LastName:    "Ivanov",
	})

	assertCode(t, err, codes.AlreadyExists)
}

// ─── Login ───────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	hashedPwd := hashPassword(t, "correctpassword")

	userRepo := &mockUserRepo{
		getUserByEmail: func(_ context.Context, _ string) (*entities.UserGetByEmail, Error.CodeError) {
			return &entities.UserGetByEmail{UserUUID: "user-uuid-1", PasswordHash: hashedPwd}, ok()
		},
	}
	authRepo := &mockAuthRepo{
		saveRefreshToken: func(_ context.Context, _, _ string) Error.CodeError { return ok() },
	}
	svc := newTestService(userRepo, authRepo)

	resp, err := svc.Login(context.Background(), &pb.LoginRequest{
		OperationId: "op-1",
		Email:       "test@example.com",
		Password:    "correctpassword",
	})

	assertNoError(t, err)
	if resp.GetAccessToken() == "" || resp.GetRefreshToken() == "" {
		t.Error("expected non-empty tokens in response")
	}
	if resp.GetUserUuid() != "user-uuid-1" {
		t.Errorf("expected user UUID 'user-uuid-1', got %q", resp.GetUserUuid())
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepo{
		getUserByEmail: func(_ context.Context, _ string) (*entities.UserGetByEmail, Error.CodeError) {
			return nil, Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("user not found")}
		},
	}
	svc := newTestService(userRepo, emptyAuthRepo())

	_, err := svc.Login(context.Background(), &pb.LoginRequest{
		OperationId: "op-1",
		Email:       "notexist@example.com",
		Password:    "password",
	})

	assertCode(t, err, codes.NotFound)
}

func TestLogin_WrongPassword(t *testing.T) {
	hashedPwd := hashPassword(t, "correctpassword")

	userRepo := &mockUserRepo{
		getUserByEmail: func(_ context.Context, _ string) (*entities.UserGetByEmail, Error.CodeError) {
			return &entities.UserGetByEmail{UserUUID: "user-uuid-1", PasswordHash: hashedPwd}, ok()
		},
	}
	svc := newTestService(userRepo, emptyAuthRepo())

	_, err := svc.Login(context.Background(), &pb.LoginRequest{
		OperationId: "op-1",
		Email:       "test@example.com",
		Password:    "wrongpassword",
	})

	assertCode(t, err, codes.InvalidArgument)
}

func TestLogin_SaveTokenError(t *testing.T) {
	hashedPwd := hashPassword(t, "correctpassword")

	userRepo := &mockUserRepo{
		getUserByEmail: func(_ context.Context, _ string) (*entities.UserGetByEmail, Error.CodeError) {
			return &entities.UserGetByEmail{UserUUID: "user-uuid-1", PasswordHash: hashedPwd}, ok()
		},
	}
	authRepo := &mockAuthRepo{
		saveRefreshToken: func(_ context.Context, _, _ string) Error.CodeError {
			return Error.CodeError{Code: int(codes.Internal), Err: fmt.Errorf("redis error")}
		},
	}
	svc := newTestService(userRepo, authRepo)

	_, err := svc.Login(context.Background(), &pb.LoginRequest{
		OperationId: "op-1",
		Email:       "test@example.com",
		Password:    "correctpassword",
	})

	assertCode(t, err, codes.Internal)
}

// ─── GetUser ─────────────────────────────────────────────────────────────────

func TestGetUser_Success(t *testing.T) {
	userRepo := &mockUserRepo{
		getUser: func(_ context.Context, uuid string) (*entities.UserGet, Error.CodeError) {
			return &entities.UserGet{
				UserUUID:  uuid,
				Email:     "test@example.com",
				FirstName: "Ivan",
				LastName:  "Ivanov",
			}, ok()
		},
	}
	svc := newTestService(userRepo, emptyAuthRepo())

	resp, err := svc.GetUser(context.Background(), &pb.GetUserRequest{
		OperationId: "op-1",
		UserUuid:    "user-uuid-1",
	})

	assertNoError(t, err)
	if resp.GetEmail() != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", resp.GetEmail())
	}
}

func TestGetUser_NotFound(t *testing.T) {
	userRepo := &mockUserRepo{
		getUser: func(_ context.Context, _ string) (*entities.UserGet, Error.CodeError) {
			return nil, Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("user not found")}
		},
	}
	svc := newTestService(userRepo, emptyAuthRepo())

	_, err := svc.GetUser(context.Background(), &pb.GetUserRequest{
		OperationId: "op-1",
		UserUuid:    "nonexistent",
	})

	assertCode(t, err, codes.NotFound)
}

// ─── ChangePassword ──────────────────────────────────────────────────────────

func TestChangePassword_Success(t *testing.T) {
	userRepo := &mockUserRepo{
		updateUserPassword: func(_ context.Context, _, _ string) Error.CodeError { return ok() },
	}
	svc := newTestService(userRepo, emptyAuthRepo())

	_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
		OperationId: "op-1",
		UserUuid:    "user-uuid-1",
		Password:    "newpassword123",
	})

	assertNoError(t, err)
}

func TestChangePassword_PasswordTooLong(t *testing.T) {
	svc := newTestService(emptyUserRepo(), emptyAuthRepo())

	_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
		OperationId: "op-1",
		UserUuid:    "user-uuid-1",
		Password:    string(make([]byte, 70)),
	})

	assertCode(t, err, codes.InvalidArgument)
}

func TestChangePassword_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepo{
		updateUserPassword: func(_ context.Context, _, _ string) Error.CodeError {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("user not found")}
		},
	}
	svc := newTestService(userRepo, emptyAuthRepo())

	_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
		OperationId: "op-1",
		UserUuid:    "nonexistent",
		Password:    "newpassword123",
	})

	assertCode(t, err, codes.NotFound)
}

// ─── UpdateUserBio ───────────────────────────────────────────────────────────

func TestUpdateUserBio_Success(t *testing.T) {
	userRepo := &mockUserRepo{
		updateUserBio: func(_ context.Context, _ *entities.UserUpdateBio) Error.CodeError { return ok() },
	}
	svc := newTestService(userRepo, emptyAuthRepo())

	_, err := svc.UpdateUserBio(context.Background(), &pb.UpdateUserBioRequest{
		OperationId: "op-1",
		UserUuid:    "user-uuid-1",
		FirstName:   "Petr",
		LastName:    "Petrov",
	})

	assertNoError(t, err)
}

func TestUpdateUserBio_NotFound(t *testing.T) {
	userRepo := &mockUserRepo{
		updateUserBio: func(_ context.Context, _ *entities.UserUpdateBio) Error.CodeError {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("user not found")}
		},
	}
	svc := newTestService(userRepo, emptyAuthRepo())

	_, err := svc.UpdateUserBio(context.Background(), &pb.UpdateUserBioRequest{
		OperationId: "op-1",
		UserUuid:    "nonexistent",
	})

	assertCode(t, err, codes.NotFound)
}

// ─── DeleteUser ──────────────────────────────────────────────────────────────

func TestDeleteUser_Success(t *testing.T) {
	userRepo := &mockUserRepo{
		deleteUser: func(_ context.Context, _ string) Error.CodeError { return ok() },
	}
	authRepo := &mockAuthRepo{
		revokeAllRefreshTokens: func(_ context.Context, _ string) Error.CodeError { return ok() },
	}
	svc := newTestService(userRepo, authRepo)

	_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{
		OperationId:       "op-1",
		InitiatorUserUuid: "user-uuid-1",
		TargetUserUuid:    "user-uuid-1",
	})

	assertNoError(t, err)
}

func TestDeleteUser_NoTokens_StillDeletes(t *testing.T) {
	// Пользователь без активных сессий всё равно должен удаляться
	userRepo := &mockUserRepo{
		deleteUser: func(_ context.Context, _ string) Error.CodeError { return ok() },
	}
	authRepo := &mockAuthRepo{
		revokeAllRefreshTokens: func(_ context.Context, _ string) Error.CodeError {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("no tokens")}
		},
	}
	svc := newTestService(userRepo, authRepo)

	_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{
		OperationId:       "op-1",
		InitiatorUserUuid: "user-uuid-1",
		TargetUserUuid:    "user-uuid-1",
	})

	assertNoError(t, err)
}

func TestDeleteUser_NotOwner(t *testing.T) {
	svc := newTestService(emptyUserRepo(), emptyAuthRepo())

	_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{
		OperationId:       "op-1",
		InitiatorUserUuid: "user-uuid-1",
		TargetUserUuid:    "user-uuid-2",
	})

	assertCode(t, err, codes.PermissionDenied)
}

func TestDeleteUser_DBError(t *testing.T) {
	userRepo := &mockUserRepo{
		deleteUser: func(_ context.Context, _ string) Error.CodeError {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("user not found")}
		},
	}
	authRepo := &mockAuthRepo{
		revokeAllRefreshTokens: func(_ context.Context, _ string) Error.CodeError { return ok() },
	}
	svc := newTestService(userRepo, authRepo)

	_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{
		OperationId:       "op-1",
		InitiatorUserUuid: "user-uuid-1",
		TargetUserUuid:    "user-uuid-1",
	})

	assertCode(t, err, codes.NotFound)
}

// ─── GetAllActiveTokens ───────────────────────────────────────────────────────

func TestGetAllActiveTokens_Success(t *testing.T) {
	authRepo := &mockAuthRepo{
		getAllRefreshTokens: func(_ context.Context, _ string) ([]string, Error.CodeError) {
			return []string{"hash1", "hash2", "hash3"}, ok()
		},
	}
	svc := newTestService(emptyUserRepo(), authRepo)

	resp, err := svc.GetAllActiveTokens(context.Background(), &pb.GetAllActiveTokensRequest{
		OperationId: "op-1",
		UserUuid:    "user-uuid-1",
	})

	assertNoError(t, err)
	if len(resp.GetTokens()) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(resp.GetTokens()))
	}
}

func TestGetAllActiveTokens_Empty(t *testing.T) {
	authRepo := &mockAuthRepo{
		getAllRefreshTokens: func(_ context.Context, _ string) ([]string, Error.CodeError) {
			return []string{}, ok()
		},
	}
	svc := newTestService(emptyUserRepo(), authRepo)

	resp, err := svc.GetAllActiveTokens(context.Background(), &pb.GetAllActiveTokensRequest{
		OperationId: "op-1",
		UserUuid:    "user-uuid-1",
	})

	assertNoError(t, err)
	if len(resp.GetTokens()) != 0 {
		t.Errorf("expected 0 tokens, got %d", len(resp.GetTokens()))
	}
}

func TestGetAllActiveTokens_CacheError(t *testing.T) {
	authRepo := &mockAuthRepo{
		getAllRefreshTokens: func(_ context.Context, _ string) ([]string, Error.CodeError) {
			return nil, Error.CodeError{Code: int(codes.Internal), Err: fmt.Errorf("redis error")}
		},
	}
	svc := newTestService(emptyUserRepo(), authRepo)

	_, err := svc.GetAllActiveTokens(context.Background(), &pb.GetAllActiveTokensRequest{
		OperationId: "op-1",
		UserUuid:    "user-uuid-1",
	})

	assertCode(t, err, codes.Internal)
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

func TestRefreshToken_Success(t *testing.T) {
	refreshToken := validRefreshToken(t)

	userRepo := &mockUserRepo{
		getUser: func(_ context.Context, _ string) (*entities.UserGet, Error.CodeError) {
			return &entities.UserGet{UserUUID: "user-uuid-1"}, ok()
		},
	}
	authRepo := &mockAuthRepo{
		checkRefreshTokenExists: func(_ context.Context, _, _ string) Error.CodeError { return ok() },
		refreshToken:            func(_ context.Context, _, _, _ string) Error.CodeError { return ok() },
	}
	svc := newTestService(userRepo, authRepo)

	resp, err := svc.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
		OperationId:  "op-1",
		RefreshToken: refreshToken,
	})

	assertNoError(t, err)
	if resp.GetAccessToken() == "" || resp.GetRefreshToken() == "" {
		t.Error("expected non-empty tokens in response")
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	svc := newTestService(emptyUserRepo(), emptyAuthRepo())

	_, err := svc.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
		OperationId:  "op-1",
		RefreshToken: "invalid.token.string",
	})

	assertCode(t, err, codes.InvalidArgument)
}

func TestRefreshToken_WrongTokenType(t *testing.T) {
	// access-токен передан вместо refresh-токена
	accessToken := validAccessToken(t)
	svc := newTestService(emptyUserRepo(), emptyAuthRepo())

	_, err := svc.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
		OperationId:  "op-1",
		RefreshToken: accessToken,
	})

	assertCode(t, err, codes.InvalidArgument)
}

func TestRefreshToken_TokenNotInCache(t *testing.T) {
	refreshToken := validRefreshToken(t)

	authRepo := &mockAuthRepo{
		checkRefreshTokenExists: func(_ context.Context, _, _ string) Error.CodeError {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("token not found")}
		},
	}
	svc := newTestService(emptyUserRepo(), authRepo)

	_, err := svc.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
		OperationId:  "op-1",
		RefreshToken: refreshToken,
	})

	assertCode(t, err, codes.NotFound)
}

func TestRefreshToken_UserNotFound(t *testing.T) {
	refreshToken := validRefreshToken(t)

	userRepo := &mockUserRepo{
		getUser: func(_ context.Context, _ string) (*entities.UserGet, Error.CodeError) {
			return nil, Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("user not found")}
		},
	}
	authRepo := &mockAuthRepo{
		checkRefreshTokenExists: func(_ context.Context, _, _ string) Error.CodeError { return ok() },
	}
	svc := newTestService(userRepo, authRepo)

	_, err := svc.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
		OperationId:  "op-1",
		RefreshToken: refreshToken,
	})

	assertCode(t, err, codes.NotFound)
}

// ─── RevokeToken ─────────────────────────────────────────────────────────────

func TestRevokeToken_Success(t *testing.T) {
	refreshToken := validRefreshToken(t)

	authRepo := &mockAuthRepo{
		revokeRefreshToken: func(_ context.Context, _, _ string) Error.CodeError { return ok() },
	}
	svc := newTestService(emptyUserRepo(), authRepo)

	_, err := svc.RevokeToken(context.Background(), &pb.RevokeTokenRequest{
		OperationId:  "op-1",
		RefreshToken: refreshToken,
	})

	assertNoError(t, err)
}

func TestRevokeToken_InvalidToken(t *testing.T) {
	svc := newTestService(emptyUserRepo(), emptyAuthRepo())

	_, err := svc.RevokeToken(context.Background(), &pb.RevokeTokenRequest{
		OperationId:  "op-1",
		RefreshToken: "bad.token",
	})

	assertCode(t, err, codes.InvalidArgument)
}

func TestRevokeToken_NotFound(t *testing.T) {
	refreshToken := validRefreshToken(t)

	authRepo := &mockAuthRepo{
		revokeRefreshToken: func(_ context.Context, _, _ string) Error.CodeError {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("token not found")}
		},
	}
	svc := newTestService(emptyUserRepo(), authRepo)

	_, err := svc.RevokeToken(context.Background(), &pb.RevokeTokenRequest{
		OperationId:  "op-1",
		RefreshToken: refreshToken,
	})

	assertCode(t, err, codes.NotFound)
}

// ─── RevokeAllTokens ─────────────────────────────────────────────────────────

func TestRevokeAllTokens_Success(t *testing.T) {
	authRepo := &mockAuthRepo{
		revokeAllRefreshTokens: func(_ context.Context, _ string) Error.CodeError { return ok() },
	}
	svc := newTestService(emptyUserRepo(), authRepo)

	_, err := svc.RevokeAllTokens(context.Background(), &pb.RevokeAllTokensRequest{
		OperationId: "op-1",
		UserUuid:    "user-uuid-1",
	})

	assertNoError(t, err)
}

func TestRevokeAllTokens_NotFound(t *testing.T) {
	authRepo := &mockAuthRepo{
		revokeAllRefreshTokens: func(_ context.Context, _ string) Error.CodeError {
			return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("no tokens")}
		},
	}
	svc := newTestService(emptyUserRepo(), authRepo)

	_, err := svc.RevokeAllTokens(context.Background(), &pb.RevokeAllTokensRequest{
		OperationId: "op-1",
		UserUuid:    "user-uuid-1",
	})

	assertCode(t, err, codes.NotFound)
}
