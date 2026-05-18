package services

import (
	"context"
	"fmt"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/utils"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

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

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hash)
}

func validRefreshToken(t *testing.T) string {
	t.Helper()
	tokens, err := utils.CreateTokens("user-uuid-1", testSecret, testAccessTTL, testRefreshTTL)
	if err != nil {
		t.Fatalf("failed to create tokens: %v", err)
	}
	return tokens.RefreshToken
}

func validAccessToken(t *testing.T) string {
	t.Helper()
	tokens, err := utils.CreateTokens("user-uuid-1", testSecret, testAccessTTL, testRefreshTTL)
	if err != nil {
		t.Fatalf("failed to create tokens: %v", err)
	}
	return tokens.AccessToken
}

// ─── Register ────────────────────────────────────────────────────────────────

func TestRegister(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			createUser: func(_ context.Context, dto *entities.UserCreate) Error.CodeError {
				if dto.Email != "test@example.com" {
					return Error.Internal(fmt.Errorf("unexpected email"))
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
	})

	t.Run("password_too_long", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.Register(context.Background(), &pb.RegisterRequest{
			OperationId: "op-1",
			Password:    string(make([]byte, 73)),
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("password_exactly_72_bytes_allowed", func(t *testing.T) {
		userRepo := &mockUserRepo{
			createUser: func(_ context.Context, _ *entities.UserCreate) Error.CodeError { return ok() },
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.Register(context.Background(), &pb.RegisterRequest{
			OperationId: "op-1",
			Email:       "test@example.com",
			Password:    string(make([]byte, 72)),
			FirstName:   "Ivan",
			LastName:    "Ivanov",
		})

		assertNoError(t, err)
	})

	t.Run("email_already_exists", func(t *testing.T) {
		userRepo := &mockUserRepo{
			createUser: func(_ context.Context, _ *entities.UserCreate) Error.CodeError {
				return Error.Public(codes.AlreadyExists, "email already registered")
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
	})
}

// ─── Login ───────────────────────────────────────────────────────────────────

func TestLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("user_not_found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ string) (*entities.UserGetByEmail, Error.CodeError) {
				return nil, Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.Login(context.Background(), &pb.LoginRequest{
			OperationId: "op-1",
			Email:       "notexist@example.com",
			Password:    "password",
		})

		assertCode(t, err, codes.NotFound)
	})

	t.Run("wrong_password", func(t *testing.T) {
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
	})

	t.Run("save_token_error", func(t *testing.T) {
		hashedPwd := hashPassword(t, "correctpassword")
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ string) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: "user-uuid-1", PasswordHash: hashedPwd}, ok()
			},
		}
		authRepo := &mockAuthRepo{
			saveRefreshToken: func(_ context.Context, _, _ string) Error.CodeError {
				return Error.Internal(fmt.Errorf("redis error"))
			},
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.Login(context.Background(), &pb.LoginRequest{
			OperationId: "op-1",
			Email:       "test@example.com",
			Password:    "correctpassword",
		})

		assertCode(t, err, codes.Internal)
	})
}

// ─── GetUser ─────────────────────────────────────────────────────────────────

func TestGetUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("not_found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUser: func(_ context.Context, _ string) (*entities.UserGet, Error.CodeError) {
				return nil, Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.GetUser(context.Background(), &pb.GetUserRequest{
			OperationId: "op-1",
			UserUuid:    "nonexistent",
		})

		assertCode(t, err, codes.NotFound)
	})
}

// ─── ChangePassword ──────────────────────────────────────────────────────────

func TestChangePassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateUserPassword: func(_ context.Context, _, _ string) Error.CodeError { return ok() },
		}
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ string) Error.CodeError { return ok() },
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
			OperationId: "op-1",
			UserUuid:    "user-uuid-1",
			Password:    "newpassword123",
		})

		assertNoError(t, err)
	})

	t.Run("success_no_active_tokens", func(t *testing.T) {
		// Если у пользователя нет токенов — ошибка NotFound из RevokeAll игнорируется
		userRepo := &mockUserRepo{
			updateUserPassword: func(_ context.Context, _, _ string) Error.CodeError { return ok() },
		}
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ string) Error.CodeError {
				return Error.Public(codes.NotFound, "no tokens")
			},
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
			OperationId: "op-1",
			UserUuid:    "user-uuid-1",
			Password:    "newpassword123",
		})

		assertNoError(t, err)
	})

	t.Run("password_too_long", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
			OperationId: "op-1",
			UserUuid:    "user-uuid-1",
			Password:    string(make([]byte, 73)),
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("user_not_found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateUserPassword: func(_ context.Context, _, _ string) Error.CodeError {
				return Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
			OperationId: "op-1",
			UserUuid:    "nonexistent",
			Password:    "newpassword123",
		})

		assertCode(t, err, codes.NotFound)
	})

	t.Run("revoke_tokens_error", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateUserPassword: func(_ context.Context, _, _ string) Error.CodeError { return ok() },
		}
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ string) Error.CodeError {
				return Error.Internal(fmt.Errorf("redis error"))
			},
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
			OperationId: "op-1",
			UserUuid:    "user-uuid-1",
			Password:    "newpassword123",
		})

		assertCode(t, err, codes.Internal)
	})
}

// ─── UpdateUserBio ───────────────────────────────────────────────────────────

func TestUpdateUserBio(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("not_found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateUserBio: func(_ context.Context, _ *entities.UserUpdateBio) Error.CodeError {
				return Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.UpdateUserBio(context.Background(), &pb.UpdateUserBioRequest{
			OperationId: "op-1",
			UserUuid:    "nonexistent",
		})

		assertCode(t, err, codes.NotFound)
	})
}

// ─── DeleteUser ──────────────────────────────────────────────────────────────

func TestDeleteUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("no_tokens_still_deletes", func(t *testing.T) {
		userRepo := &mockUserRepo{
			deleteUser: func(_ context.Context, _ string) Error.CodeError { return ok() },
		}
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ string) Error.CodeError {
				return Error.Public(codes.NotFound, "no tokens")
			},
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{
			OperationId:       "op-1",
			InitiatorUserUuid: "user-uuid-1",
			TargetUserUuid:    "user-uuid-1",
		})

		assertNoError(t, err)
	})

	t.Run("not_owner", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{
			OperationId:       "op-1",
			InitiatorUserUuid: "user-uuid-1",
			TargetUserUuid:    "user-uuid-2",
		})

		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("user_not_found_in_db", func(t *testing.T) {
		userRepo := &mockUserRepo{
			deleteUser: func(_ context.Context, _ string) Error.CodeError {
				return Error.Public(codes.NotFound, "user not found")
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
	})
}

// ─── GetAllActiveTokens ───────────────────────────────────────────────────────

func TestGetAllActiveTokens(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("empty", func(t *testing.T) {
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
	})

	t.Run("cache_error", func(t *testing.T) {
		authRepo := &mockAuthRepo{
			getAllRefreshTokens: func(_ context.Context, _ string) ([]string, Error.CodeError) {
				return nil, Error.Internal(fmt.Errorf("redis error"))
			},
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.GetAllActiveTokens(context.Background(), &pb.GetAllActiveTokensRequest{
			OperationId: "op-1",
			UserUuid:    "user-uuid-1",
		})

		assertCode(t, err, codes.Internal)
	})
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

func TestRefreshToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("invalid_token", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
			OperationId:  "op-1",
			RefreshToken: "invalid.token.string",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("wrong_token_type", func(t *testing.T) {
		accessToken := validAccessToken(t)
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
			OperationId:  "op-1",
			RefreshToken: accessToken,
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("token_not_in_cache", func(t *testing.T) {
		refreshToken := validRefreshToken(t)
		authRepo := &mockAuthRepo{
			checkRefreshTokenExists: func(_ context.Context, _, _ string) Error.CodeError {
				return Error.Public(codes.NotFound, "token not found")
			},
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
			OperationId:  "op-1",
			RefreshToken: refreshToken,
		})

		assertCode(t, err, codes.NotFound)
	})

	t.Run("user_not_found", func(t *testing.T) {
		refreshToken := validRefreshToken(t)
		userRepo := &mockUserRepo{
			getUser: func(_ context.Context, _ string) (*entities.UserGet, Error.CodeError) {
				return nil, Error.Public(codes.NotFound, "user not found")
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
	})
}

// ─── RevokeToken ─────────────────────────────────────────────────────────────

func TestRevokeToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		authRepo := &mockAuthRepo{
			revokeRefreshToken: func(_ context.Context, _, _ string) Error.CodeError { return ok() },
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.RevokeToken(context.Background(), &pb.RevokeTokenRequest{
			OperationId: "op-1",
			UserUuid:    "user-uuid-1",
			TokenHash:   "abc123hash",
		})

		assertNoError(t, err)
	})

	t.Run("token_not_found", func(t *testing.T) {
		authRepo := &mockAuthRepo{
			revokeRefreshToken: func(_ context.Context, _, _ string) Error.CodeError {
				return Error.Public(codes.NotFound, "refresh token not found")
			},
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.RevokeToken(context.Background(), &pb.RevokeTokenRequest{
			OperationId: "op-1",
			UserUuid:    "user-uuid-1",
			TokenHash:   "abc123hash",
		})

		assertCode(t, err, codes.NotFound)
	})

	t.Run("token_belongs_to_different_user", func(t *testing.T) {
		// Слой репозитория проверяет принадлежность и возвращает NotFound
		authRepo := &mockAuthRepo{
			revokeRefreshToken: func(_ context.Context, userUUID, _ string) Error.CodeError {
				if userUUID != "owner-uuid" {
					return Error.Public(codes.NotFound, "refresh token not found")
				}
				return ok()
			},
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.RevokeToken(context.Background(), &pb.RevokeTokenRequest{
			OperationId: "op-1",
			UserUuid:    "attacker-uuid",
			TokenHash:   "abc123hash",
		})

		assertCode(t, err, codes.NotFound)
	})
}

// ─── RevokeAllTokens ─────────────────────────────────────────────────────────

func TestRevokeAllTokens(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ string) Error.CodeError { return ok() },
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.RevokeAllTokens(context.Background(), &pb.RevokeAllTokensRequest{
			OperationId: "op-1",
			UserUuid:    "user-uuid-1",
		})

		assertNoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ string) Error.CodeError {
				return Error.Public(codes.NotFound, "no tokens")
			},
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.RevokeAllTokens(context.Background(), &pb.RevokeAllTokensRequest{
			OperationId: "op-1",
			UserUuid:    "user-uuid-1",
		})

		assertCode(t, err, codes.NotFound)
	})
}
