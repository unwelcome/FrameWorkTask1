package services

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/entities"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/utils"
	pb "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
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
	tokens, err := utils.CreateTokens(testUUID1, testPrivateKey, testAccessTTL, testRefreshTTL)
	if err != nil {
		t.Fatalf("failed to create tokens: %v", err)
	}
	return tokens.RefreshToken
}

func validAccessToken(t *testing.T) string {
	t.Helper()
	tokens, err := utils.CreateTokens(testUUID1, testPrivateKey, testAccessTTL, testRefreshTTL)
	if err != nil {
		t.Fatalf("failed to create tokens: %v", err)
	}
	return tokens.AccessToken
}

// maxLenPassword возвращает валидный пароль длиной ровно 72 байта.
// Содержит uppercase, lowercase и цифру — проходит все проверки.
func maxLenPassword() string {
	return "Password1" + strings.Repeat("a", 63) // 9 + 63 = 72 байта
}

// ─── Register ────────────────────────────────────────────────────────────────

func TestRegister(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			createUser: func(_ context.Context, dto entities.User) Error.CodeError {
				if dto.Email != "test@example.com" {
					return Error.Internal(fmt.Errorf("unexpected email"))
				}
				return ok()
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		resp, err := svc.Register(context.Background(), &pb.RegisterRequest{
			Email:     "test@example.com",
			Password:  testPassword,
			FirstName: "Ivan",
			LastName:  "Ivanov",
		})

		assertNoError(t, err)
		if resp.GetUserUuid() == "" {
			t.Error("expected non-empty user UUID")
		}
	})

	t.Run("invalid_email", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.Register(context.Background(), &pb.RegisterRequest{
			Email:     "not-an-email",
			Password:  testPassword,
			FirstName: "Ivan",
			LastName:  "Ivanov",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("password_too_short", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.Register(context.Background(), &pb.RegisterRequest{
			Email:    "test@example.com",
			Password: "Ab1",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("password_too_long", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.Register(context.Background(), &pb.RegisterRequest{
			Email:    "test@example.com",
			Password: "Password1" + strings.Repeat("a", 64), // 73 байта
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("password_exactly_72_bytes_allowed", func(t *testing.T) {
		userRepo := &mockUserRepo{
			createUser: func(_ context.Context, _ entities.User) Error.CodeError { return ok() },
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.Register(context.Background(), &pb.RegisterRequest{
			Email:     "test@example.com",
			Password:  maxLenPassword(),
			FirstName: "Ivan",
			LastName:  "Ivanov",
		})

		assertNoError(t, err)
	})

	t.Run("invalid_first_name", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.Register(context.Background(), &pb.RegisterRequest{
			Email:     "test@example.com",
			Password:  testPassword,
			FirstName: "Ivan123",
			LastName:  "Ivanov",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_last_name", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.Register(context.Background(), &pb.RegisterRequest{
			Email:     "test@example.com",
			Password:  testPassword,
			FirstName: "Ivan",
			LastName:  "Ivanov123",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("email_already_exists", func(t *testing.T) {
		userRepo := &mockUserRepo{
			createUser: func(_ context.Context, _ entities.User) Error.CodeError {
				return Error.Public(codes.AlreadyExists, "email already registered")
			},
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID2, IsVerified: true}, ok()
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.Register(context.Background(), &pb.RegisterRequest{
			Email:     "taken@example.com",
			Password:  testPassword,
			FirstName: "Ivan",
			LastName:  "Ivanov",
		})

		assertCode(t, err, codes.AlreadyExists)
	})
}

// ─── Login ───────────────────────────────────────────────────────────────────

func TestLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		hashedPwd := hashPassword(t, testPassword)
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, PasswordHash: hashedPwd, IsVerified: true}, ok()
			},
		}
		authRepo := &mockAuthRepo{
			saveRefreshToken: func(_ context.Context, _ entities.SaveRefreshTokenDTO) Error.CodeError { return ok() },
		}
		svc := newTestService(userRepo, authRepo)

		resp, err := svc.Login(context.Background(), &pb.LoginRequest{
			Email:    "test@example.com",
			Password: testPassword,
		})

		assertNoError(t, err)
		if resp.GetAccessToken() == "" || resp.GetRefreshToken() == "" {
			t.Error("expected non-empty tokens in response")
		}
		if resp.GetUserUuid() != testUUID1 {
			t.Errorf("expected user UUID %q, got %q", testUUID1, resp.GetUserUuid())
		}
	})

	t.Run("invalid_email", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.Login(context.Background(), &pb.LoginRequest{
			Email:    "not-an-email",
			Password: testPassword,
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_password", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.Login(context.Background(), &pb.LoginRequest{
			Email:    "test@example.com",
			Password: "short",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("user_not_found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return nil, Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.Login(context.Background(), &pb.LoginRequest{
			Email:    "notexist@example.com",
			Password: testPassword,
		})

		// После исправления timing-атаки сервис запускает bcrypt на фиктивном хеше
		// и возвращает тот же код, что и при неверном пароле — чтобы не раскрывать
		// существование email по разнице в ~200 мс или по HTTP-статусу.
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("wrong_password", func(t *testing.T) {
		hashedPwd := hashPassword(t, testPassword)
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, PasswordHash: hashedPwd}, ok()
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.Login(context.Background(), &pb.LoginRequest{
			Email:    "test@example.com",
			Password: "WrongPassword1",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("save_token_error", func(t *testing.T) {
		hashedPwd := hashPassword(t, testPassword)
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, PasswordHash: hashedPwd, IsVerified: true}, ok()
			},
		}
		authRepo := &mockAuthRepo{
			saveRefreshToken: func(_ context.Context, _ entities.SaveRefreshTokenDTO) Error.CodeError {
				return Error.Internal(fmt.Errorf("redis error"))
			},
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.Login(context.Background(), &pb.LoginRequest{
			Email:    "test@example.com",
			Password: testPassword,
		})

		assertCode(t, err, codes.Internal)
	})
}

// ─── GetUser ─────────────────────────────────────────────────────────────────

func TestGetUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUser: func(_ context.Context, dto entities.GetUserDTO) (*entities.UserGet, Error.CodeError) {
				return &entities.UserGet{
					UserUUID:  dto.UserUUID,
					Email:     "test@example.com",
					FirstName: "Ivan",
					LastName:  "Ivanov",
				}, ok()
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		resp, err := svc.GetUser(context.Background(), &pb.GetUserRequest{
			UserUuid: testUUID1,
		})

		assertNoError(t, err)
		if resp.GetEmail() != "test@example.com" {
			t.Errorf("expected email 'test@example.com', got %q", resp.GetEmail())
		}
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.GetUser(context.Background(), &pb.GetUserRequest{
			UserUuid: "not-a-uuid",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("not_found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUser: func(_ context.Context, _ entities.GetUserDTO) (*entities.UserGet, Error.CodeError) {
				return nil, Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.GetUser(context.Background(), &pb.GetUserRequest{
			UserUuid: testUUID1,
		})

		assertCode(t, err, codes.NotFound)
	})
}

// ─── ChangePassword ──────────────────────────────────────────────────────────

func TestChangePassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateUserPassword: func(_ context.Context, _ entities.UpdateUserPasswordDTO) Error.CodeError { return ok() },
		}
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ entities.RevokeAllRefreshTokensDTO) Error.CodeError { return ok() },
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
			UserUuid: testUUID1,
			Password: testPassword,
		})

		assertNoError(t, err)
	})

	t.Run("success_no_active_tokens", func(t *testing.T) {
		// Если у пользователя нет токенов — ошибка NotFound из RevokeAll игнорируется
		userRepo := &mockUserRepo{
			updateUserPassword: func(_ context.Context, _ entities.UpdateUserPasswordDTO) Error.CodeError { return ok() },
		}
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ entities.RevokeAllRefreshTokensDTO) Error.CodeError {
				return Error.Public(codes.NotFound, "no tokens")
			},
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
			UserUuid: testUUID1,
			Password: testPassword,
		})

		assertNoError(t, err)
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
			UserUuid: "not-a-uuid",
			Password: testPassword,
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("password_too_short", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
			UserUuid: testUUID1,
			Password: "Ab1",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("password_too_long", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
			UserUuid: testUUID1,
			Password: "Password1" + strings.Repeat("a", 64), // 73 байта
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("user_not_found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateUserPassword: func(_ context.Context, _ entities.UpdateUserPasswordDTO) Error.CodeError {
				return Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
			UserUuid: testUUID1,
			Password: testPassword,
		})

		assertCode(t, err, codes.NotFound)
	})

	t.Run("revoke_tokens_error", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateUserPassword: func(_ context.Context, _ entities.UpdateUserPasswordDTO) Error.CodeError { return ok() },
		}
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ entities.RevokeAllRefreshTokensDTO) Error.CodeError {
				return Error.Internal(fmt.Errorf("redis error"))
			},
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.ChangePassword(context.Background(), &pb.ChangePasswordRequest{
			UserUuid: testUUID1,
			Password: testPassword,
		})

		assertCode(t, err, codes.Internal)
	})
}

// ─── UpdateUserBio ───────────────────────────────────────────────────────────

func TestUpdateUserBio(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateUserBio: func(_ context.Context, _ entities.UserUpdateBioDTO) Error.CodeError { return ok() },
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.UpdateUserBio(context.Background(), &pb.UpdateUserBioRequest{
			UserUuid:  testUUID1,
			FirstName: "Petr",
			LastName:  "Petrov",
		})

		assertNoError(t, err)
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.UpdateUserBio(context.Background(), &pb.UpdateUserBioRequest{
			UserUuid:  "not-a-uuid",
			FirstName: "Petr",
			LastName:  "Petrov",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_first_name", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.UpdateUserBio(context.Background(), &pb.UpdateUserBioRequest{
			UserUuid:  testUUID1,
			FirstName: "Petr123",
			LastName:  "Petrov",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("not_found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateUserBio: func(_ context.Context, _ entities.UserUpdateBioDTO) Error.CodeError {
				return Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.UpdateUserBio(context.Background(), &pb.UpdateUserBioRequest{
			UserUuid:  testUUID1,
			FirstName: "Petr",
			LastName:  "Petrov",
		})

		assertCode(t, err, codes.NotFound)
	})
}

// ─── DeleteUser ──────────────────────────────────────────────────────────────

func TestDeleteUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			deleteUser: func(_ context.Context, _ entities.DeleteUserDTO) Error.CodeError { return ok() },
		}
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ entities.RevokeAllRefreshTokensDTO) Error.CodeError { return ok() },
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{
			InitiatorUserUuid: testUUID1,
			TargetUserUuid:    testUUID1,
		})

		assertNoError(t, err)
	})

	t.Run("no_tokens_still_deletes", func(t *testing.T) {
		userRepo := &mockUserRepo{
			deleteUser: func(_ context.Context, _ entities.DeleteUserDTO) Error.CodeError { return ok() },
		}
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ entities.RevokeAllRefreshTokensDTO) Error.CodeError {
				return Error.Public(codes.NotFound, "no tokens")
			},
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{
			InitiatorUserUuid: testUUID1,
			TargetUserUuid:    testUUID1,
		})

		assertNoError(t, err)
	})

	t.Run("invalid_initiator_uuid", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{
			InitiatorUserUuid: "not-a-uuid",
			TargetUserUuid:    testUUID1,
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_target_uuid", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{
			InitiatorUserUuid: testUUID1,
			TargetUserUuid:    "not-a-uuid",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("not_owner", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{
			InitiatorUserUuid: testUUID1,
			TargetUserUuid:    testUUID2,
		})

		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("user_not_found_in_db", func(t *testing.T) {
		userRepo := &mockUserRepo{
			deleteUser: func(_ context.Context, _ entities.DeleteUserDTO) Error.CodeError {
				return Error.Public(codes.NotFound, "user not found")
			},
		}
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ entities.RevokeAllRefreshTokensDTO) Error.CodeError { return ok() },
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{
			InitiatorUserUuid: testUUID1,
			TargetUserUuid:    testUUID1,
		})

		assertCode(t, err, codes.NotFound)
	})
}

// ─── GetAllActiveTokens ───────────────────────────────────────────────────────

func TestGetAllActiveTokens(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		authRepo := &mockAuthRepo{
			getAllRefreshTokens: func(_ context.Context, _ entities.GetAllRefreshTokensDTO) ([]string, Error.CodeError) {
				return []string{"hash1", "hash2", "hash3"}, ok()
			},
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		resp, err := svc.GetAllActiveTokens(context.Background(), &pb.GetAllActiveTokensRequest{
			UserUuid: testUUID1,
		})

		assertNoError(t, err)
		if len(resp.GetTokens()) != 3 {
			t.Errorf("expected 3 tokens, got %d", len(resp.GetTokens()))
		}
	})

	t.Run("empty", func(t *testing.T) {
		authRepo := &mockAuthRepo{
			getAllRefreshTokens: func(_ context.Context, _ entities.GetAllRefreshTokensDTO) ([]string, Error.CodeError) {
				return []string{}, ok()
			},
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		resp, err := svc.GetAllActiveTokens(context.Background(), &pb.GetAllActiveTokensRequest{
			UserUuid: testUUID1,
		})

		assertNoError(t, err)
		if len(resp.GetTokens()) != 0 {
			t.Errorf("expected 0 tokens, got %d", len(resp.GetTokens()))
		}
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.GetAllActiveTokens(context.Background(), &pb.GetAllActiveTokensRequest{
			UserUuid: "not-a-uuid",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("cache_error", func(t *testing.T) {
		authRepo := &mockAuthRepo{
			getAllRefreshTokens: func(_ context.Context, _ entities.GetAllRefreshTokensDTO) ([]string, Error.CodeError) {
				return nil, Error.Internal(fmt.Errorf("redis error"))
			},
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.GetAllActiveTokens(context.Background(), &pb.GetAllActiveTokensRequest{
			UserUuid: testUUID1,
		})

		assertCode(t, err, codes.Internal)
	})
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

func TestRefreshToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		refreshToken := validRefreshToken(t)
		userRepo := &mockUserRepo{
			getUser: func(_ context.Context, _ entities.GetUserDTO) (*entities.UserGet, Error.CodeError) {
				return &entities.UserGet{UserUUID: testUUID1}, ok()
			},
		}
		authRepo := &mockAuthRepo{
			checkRefreshTokenExists: func(_ context.Context, _ entities.CheckRefreshTokenExistsDTO) Error.CodeError { return ok() },
			refreshToken:            func(_ context.Context, _ entities.RefreshTokenDTO) Error.CodeError { return ok() },
		}
		svc := newTestService(userRepo, authRepo)

		resp, err := svc.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
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
			RefreshToken: "invalid.token.string",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("wrong_token_type", func(t *testing.T) {
		accessToken := validAccessToken(t)
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
			RefreshToken: accessToken,
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("token_not_in_cache", func(t *testing.T) {
		refreshToken := validRefreshToken(t)
		authRepo := &mockAuthRepo{
			checkRefreshTokenExists: func(_ context.Context, _ entities.CheckRefreshTokenExistsDTO) Error.CodeError {
				return Error.Public(codes.NotFound, "token not found")
			},
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
			RefreshToken: refreshToken,
		})

		assertCode(t, err, codes.NotFound)
	})

	t.Run("user_not_found", func(t *testing.T) {
		refreshToken := validRefreshToken(t)
		userRepo := &mockUserRepo{
			getUser: func(_ context.Context, _ entities.GetUserDTO) (*entities.UserGet, Error.CodeError) {
				return nil, Error.Public(codes.NotFound, "user not found")
			},
		}
		authRepo := &mockAuthRepo{
			checkRefreshTokenExists: func(_ context.Context, _ entities.CheckRefreshTokenExistsDTO) Error.CodeError { return ok() },
		}
		svc := newTestService(userRepo, authRepo)

		_, err := svc.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
			RefreshToken: refreshToken,
		})

		assertCode(t, err, codes.NotFound)
	})
}

// ─── RevokeToken ─────────────────────────────────────────────────────────────

func TestRevokeToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		authRepo := &mockAuthRepo{
			revokeRefreshToken: func(_ context.Context, _ entities.RevokeRefreshTokenDTO) Error.CodeError { return ok() },
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.RevokeToken(context.Background(), &pb.RevokeTokenRequest{
			UserUuid:  testUUID1,
			TokenHash: "abc123hash",
		})

		assertNoError(t, err)
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.RevokeToken(context.Background(), &pb.RevokeTokenRequest{
			UserUuid:  "not-a-uuid",
			TokenHash: "abc123hash",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("empty_token_hash", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.RevokeToken(context.Background(), &pb.RevokeTokenRequest{
			UserUuid:  testUUID1,
			TokenHash: "",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("token_not_found", func(t *testing.T) {
		authRepo := &mockAuthRepo{
			revokeRefreshToken: func(_ context.Context, _ entities.RevokeRefreshTokenDTO) Error.CodeError {
				return Error.Public(codes.NotFound, "refresh token not found")
			},
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.RevokeToken(context.Background(), &pb.RevokeTokenRequest{
			UserUuid:  testUUID1,
			TokenHash: "abc123hash",
		})

		assertCode(t, err, codes.NotFound)
	})

	t.Run("token_belongs_to_different_user", func(t *testing.T) {
		// Слой репозитория проверяет принадлежность и возвращает NotFound
		const ownerUUID = "cccccccc-cccc-cccc-cccc-cccccccccccc"
		authRepo := &mockAuthRepo{
			revokeRefreshToken: func(_ context.Context, dto entities.RevokeRefreshTokenDTO) Error.CodeError {
				if dto.UserUUID != ownerUUID {
					return Error.Public(codes.NotFound, "refresh token not found")
				}
				return ok()
			},
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.RevokeToken(context.Background(), &pb.RevokeTokenRequest{
			UserUuid:  testUUID1, // не владелец токена
			TokenHash: "abc123hash",
		})

		assertCode(t, err, codes.NotFound)
	})
}

// ─── Login (дополнительные ветки) ────────────────────────────────────────────

func TestLogin_Extra(t *testing.T) {
	t.Run("account_not_verified", func(t *testing.T) {
		hashedPwd := hashPassword(t, testPassword)
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, PasswordHash: hashedPwd, IsVerified: false}, ok()
			},
		}
		svc := newTestService(userRepo, emptyAuthRepo())

		_, err := svc.Login(context.Background(), &pb.LoginRequest{
			Email:    "test@example.com",
			Password: testPassword,
		})

		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("two_fa_enabled_returns_session", func(t *testing.T) {
		hashedPwd := hashPassword(t, testPassword)
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{
					UserUUID:     testUUID1,
					PasswordHash: hashedPwd,
					IsVerified:   true,
					Enabled2FA:   true,
					FirstName:    "Ivan",
				}, ok()
			},
		}
		twoFARepo := &mockTwoFARepo{
			save2FAData: func(_ context.Context, _ entities.Save2FADataDTO) Error.CodeError { return ok() },
		}
		svc := buildSvc(svcDeps{user: userRepo, twoFA: twoFARepo})

		resp, err := svc.Login(context.Background(), &pb.LoginRequest{
			Email:    "test@example.com",
			Password: testPassword,
		})

		assertNoError(t, err)
		if resp.GetSessionUuid() == "" {
			t.Error("expected non-empty session_uuid when 2FA is enabled")
		}
		if resp.GetAccessToken() != "" || resp.GetRefreshToken() != "" {
			t.Error("expected empty tokens when 2FA is enabled")
		}
	})

	t.Run("two_fa_save_error", func(t *testing.T) {
		hashedPwd := hashPassword(t, testPassword)
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{
					UserUUID:     testUUID1,
					PasswordHash: hashedPwd,
					IsVerified:   true,
					Enabled2FA:   true,
				}, ok()
			},
		}
		twoFARepo := &mockTwoFARepo{
			save2FAData: func(_ context.Context, _ entities.Save2FADataDTO) Error.CodeError {
				return Error.Internal(fmt.Errorf("redis error"))
			},
		}
		svc := buildSvc(svcDeps{user: userRepo, twoFA: twoFARepo})

		_, err := svc.Login(context.Background(), &pb.LoginRequest{
			Email:    "test@example.com",
			Password: testPassword,
		})

		assertCode(t, err, codes.Internal)
	})
}

// ─── RevokeAllTokens ─────────────────────────────────────────────────────────

func TestRevokeAllTokens(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ entities.RevokeAllRefreshTokensDTO) Error.CodeError { return ok() },
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.RevokeAllTokens(context.Background(), &pb.RevokeAllTokensRequest{
			UserUuid: testUUID1,
		})

		assertNoError(t, err)
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		svc := newTestService(emptyUserRepo(), emptyAuthRepo())

		_, err := svc.RevokeAllTokens(context.Background(), &pb.RevokeAllTokensRequest{
			UserUuid: "not-a-uuid",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("not_found", func(t *testing.T) {
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ entities.RevokeAllRefreshTokensDTO) Error.CodeError {
				return Error.Public(codes.NotFound, "no tokens")
			},
		}
		svc := newTestService(emptyUserRepo(), authRepo)

		_, err := svc.RevokeAllTokens(context.Background(), &pb.RevokeAllTokensRequest{
			UserUuid: testUUID1,
		})

		assertCode(t, err, codes.NotFound)
	})
}

// ─── VerifyAccount ────────────────────────────────────────────────────────────

func TestVerifyAccount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: false}, ok()
			},
			setUserVerified: func(_ context.Context, _ entities.SetUserVerifiedDTO) Error.CodeError { return ok() },
		}
		verRepo := &mockVerificationRepo{
			getVerificationCode: func(_ context.Context, _ entities.GetVerificationCodeDTO) (string, Error.CodeError) {
				return "123456", ok()
			},
			incrVerificationAttempts: func(_ context.Context, _ entities.IncrVerificationAttemptsDTO) (int64, Error.CodeError) {
				return 1, ok()
			},
			deleteVerificationCode: func(_ context.Context, _ entities.DeleteVerificationCodeDTO) Error.CodeError { return ok() },
		}
		svc := buildSvc(svcDeps{user: userRepo, verification: verRepo})

		_, err := svc.VerifyAccount(context.Background(), &pb.VerifyAccountRequest{
			Email: "test@example.com",
			Code:  "123456",
		})

		assertNoError(t, err)
	})

	t.Run("invalid_email", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.VerifyAccount(context.Background(), &pb.VerifyAccountRequest{
			Email: "not-an-email",
			Code:  "123456",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_code_format", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.VerifyAccount(context.Background(), &pb.VerifyAccountRequest{
			Email: "test@example.com",
			Code:  "abc",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("user_not_found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return nil, Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := buildSvc(svcDeps{user: userRepo})

		_, err := svc.VerifyAccount(context.Background(), &pb.VerifyAccountRequest{
			Email: "unknown@example.com",
			Code:  "123456",
		})

		// Сервис возвращает нейтральный InvalidArgument, не раскрывая отсутствие пользователя
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("already_verified", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: true}, ok()
			},
		}
		svc := buildSvc(svcDeps{user: userRepo})

		_, err := svc.VerifyAccount(context.Background(), &pb.VerifyAccountRequest{
			Email: "test@example.com",
			Code:  "123456",
		})

		assertCode(t, err, codes.AlreadyExists)
	})

	t.Run("code_not_found_or_expired", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: false}, ok()
			},
		}
		verRepo := &mockVerificationRepo{
			getVerificationCode: func(_ context.Context, _ entities.GetVerificationCodeDTO) (string, Error.CodeError) {
				return "", Error.Public(codes.NotFound, "code not found")
			},
		}
		svc := buildSvc(svcDeps{user: userRepo, verification: verRepo})

		_, err := svc.VerifyAccount(context.Background(), &pb.VerifyAccountRequest{
			Email: "test@example.com",
			Code:  "123456",
		})

		// Код не найден/истёк должен возвращать InvalidArgument (400), а не NotFound (404),
		// чтобы атакующий не мог по HTTP-статусу определить наличие активного кода для email.
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("too_many_attempts", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: false}, ok()
			},
		}
		verRepo := &mockVerificationRepo{
			getVerificationCode: func(_ context.Context, _ entities.GetVerificationCodeDTO) (string, Error.CodeError) {
				return "123456", ok()
			},
			incrVerificationAttempts: func(_ context.Context, _ entities.IncrVerificationAttemptsDTO) (int64, Error.CodeError) {
				return maxVerificationAttempts + 1, ok()
			},
			deleteVerificationCode: func(_ context.Context, _ entities.DeleteVerificationCodeDTO) Error.CodeError { return ok() },
		}
		svc := buildSvc(svcDeps{user: userRepo, verification: verRepo})

		_, err := svc.VerifyAccount(context.Background(), &pb.VerifyAccountRequest{
			Email: "test@example.com",
			Code:  "123456",
		})

		assertCode(t, err, codes.ResourceExhausted)
	})

	t.Run("wrong_code", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: false}, ok()
			},
		}
		verRepo := &mockVerificationRepo{
			getVerificationCode: func(_ context.Context, _ entities.GetVerificationCodeDTO) (string, Error.CodeError) {
				return "999999", ok()
			},
			incrVerificationAttempts: func(_ context.Context, _ entities.IncrVerificationAttemptsDTO) (int64, Error.CodeError) {
				return 1, ok()
			},
		}
		svc := buildSvc(svcDeps{user: userRepo, verification: verRepo})

		_, err := svc.VerifyAccount(context.Background(), &pb.VerifyAccountRequest{
			Email: "test@example.com",
			Code:  "123456",
		})

		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── ResendVerificationCode ───────────────────────────────────────────────────

func TestResendVerificationCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: false, FirstName: "Ivan"}, ok()
			},
		}
		verRepo := &mockVerificationRepo{
			saveVerificationCode: func(_ context.Context, _ entities.SaveVerificationCodeDTO) Error.CodeError { return ok() },
		}
		svc := buildSvc(svcDeps{user: userRepo, verification: verRepo})

		_, err := svc.ResendVerificationCode(context.Background(), &pb.ResendVerificationCodeRequest{
			Email: "test@example.com",
		})

		assertNoError(t, err)
	})

	t.Run("invalid_email", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.ResendVerificationCode(context.Background(), &pb.ResendVerificationCodeRequest{
			Email: "not-an-email",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("user_not_found_silent_ok", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return nil, Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := buildSvc(svcDeps{user: userRepo})

		_, err := svc.ResendVerificationCode(context.Background(), &pb.ResendVerificationCodeRequest{
			Email: "unknown@example.com",
		})

		// Тихий 200 — не раскрываем наличие/отсутствие email
		assertNoError(t, err)
	})

	t.Run("already_verified", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: true}, ok()
			},
		}
		svc := buildSvc(svcDeps{user: userRepo})

		_, err := svc.ResendVerificationCode(context.Background(), &pb.ResendVerificationCodeRequest{
			Email: "test@example.com",
		})

		assertCode(t, err, codes.AlreadyExists)
	})

	t.Run("save_code_error", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: false}, ok()
			},
		}
		verRepo := &mockVerificationRepo{
			saveVerificationCode: func(_ context.Context, _ entities.SaveVerificationCodeDTO) Error.CodeError {
				return Error.Internal(fmt.Errorf("redis error"))
			},
		}
		svc := buildSvc(svcDeps{user: userRepo, verification: verRepo})

		_, err := svc.ResendVerificationCode(context.Background(), &pb.ResendVerificationCodeRequest{
			Email: "test@example.com",
		})

		assertCode(t, err, codes.Internal)
	})
}

// ─── ForgotPassword ───────────────────────────────────────────────────────────

func TestForgotPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: true, FirstName: "Ivan"}, ok()
			},
		}
		recRepo := &mockRecoveryRepo{
			saveRecoveryCode: func(_ context.Context, _ entities.SaveRecoveryCodeDTO) Error.CodeError { return ok() },
		}
		svc := buildSvc(svcDeps{user: userRepo, recovery: recRepo})

		_, err := svc.ForgotPassword(context.Background(), &pb.ForgotPasswordRequest{
			Email: "test@example.com",
		})

		assertNoError(t, err)
	})

	t.Run("invalid_email", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.ForgotPassword(context.Background(), &pb.ForgotPasswordRequest{
			Email: "not-an-email",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("user_not_found_silent_ok", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return nil, Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := buildSvc(svcDeps{user: userRepo})

		_, err := svc.ForgotPassword(context.Background(), &pb.ForgotPasswordRequest{
			Email: "unknown@example.com",
		})

		assertNoError(t, err)
	})

	t.Run("user_not_verified_silent_ok", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: false}, ok()
			},
		}
		svc := buildSvc(svcDeps{user: userRepo})

		_, err := svc.ForgotPassword(context.Background(), &pb.ForgotPasswordRequest{
			Email: "unverified@example.com",
		})

		assertNoError(t, err)
	})

	t.Run("save_code_error", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: true}, ok()
			},
		}
		recRepo := &mockRecoveryRepo{
			saveRecoveryCode: func(_ context.Context, _ entities.SaveRecoveryCodeDTO) Error.CodeError {
				return Error.Internal(fmt.Errorf("redis error"))
			},
		}
		svc := buildSvc(svcDeps{user: userRepo, recovery: recRepo})

		_, err := svc.ForgotPassword(context.Background(), &pb.ForgotPasswordRequest{
			Email: "test@example.com",
		})

		assertCode(t, err, codes.Internal)
	})
}

// ─── ResetPassword ────────────────────────────────────────────────────────────

func TestResetPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: true}, ok()
			},
			updateUserPassword: func(_ context.Context, _ entities.UpdateUserPasswordDTO) Error.CodeError { return ok() },
		}
		recRepo := &mockRecoveryRepo{
			getRecoveryCode: func(_ context.Context, _ entities.GetRecoveryCodeDTO) (string, Error.CodeError) {
				return "123456", ok()
			},
			incrRecoveryAttempts: func(_ context.Context, _ entities.IncrRecoveryAttemptsDTO) (int64, Error.CodeError) {
				return 1, ok()
			},
			deleteRecoveryCode: func(_ context.Context, _ entities.DeleteRecoveryCodeDTO) Error.CodeError { return ok() },
		}
		authRepo := &mockAuthRepo{
			revokeAllRefreshTokens: func(_ context.Context, _ entities.RevokeAllRefreshTokensDTO) Error.CodeError { return ok() },
		}
		svc := buildSvc(svcDeps{user: userRepo, auth: authRepo, recovery: recRepo})

		_, err := svc.ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			Email:       "test@example.com",
			Code:        "123456",
			NewPassword: testPassword,
		})

		assertNoError(t, err)
	})

	t.Run("invalid_email", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			Email:       "not-an-email",
			Code:        "123456",
			NewPassword: testPassword,
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_code_format", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			Email:       "test@example.com",
			Code:        "abc",
			NewPassword: testPassword,
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_new_password", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			Email:       "test@example.com",
			Code:        "123456",
			NewPassword: "weak",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("user_not_found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return nil, Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := buildSvc(svcDeps{user: userRepo})

		_, err := svc.ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			Email:       "unknown@example.com",
			Code:        "123456",
			NewPassword: testPassword,
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("user_not_verified", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: false}, ok()
			},
		}
		svc := buildSvc(svcDeps{user: userRepo})

		_, err := svc.ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			Email:       "test@example.com",
			Code:        "123456",
			NewPassword: testPassword,
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("code_not_found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: true}, ok()
			},
		}
		recRepo := &mockRecoveryRepo{
			getRecoveryCode: func(_ context.Context, _ entities.GetRecoveryCodeDTO) (string, Error.CodeError) {
				return "", Error.Public(codes.NotFound, "code not found")
			},
		}
		svc := buildSvc(svcDeps{user: userRepo, recovery: recRepo})

		_, err := svc.ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			Email:       "test@example.com",
			Code:        "123456",
			NewPassword: testPassword,
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("too_many_attempts", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: true}, ok()
			},
		}
		recRepo := &mockRecoveryRepo{
			getRecoveryCode: func(_ context.Context, _ entities.GetRecoveryCodeDTO) (string, Error.CodeError) {
				return "123456", ok()
			},
			incrRecoveryAttempts: func(_ context.Context, _ entities.IncrRecoveryAttemptsDTO) (int64, Error.CodeError) {
				return maxRecoveryAttempts + 1, ok()
			},
			deleteRecoveryCode: func(_ context.Context, _ entities.DeleteRecoveryCodeDTO) Error.CodeError { return ok() },
		}
		svc := buildSvc(svcDeps{user: userRepo, recovery: recRepo})

		_, err := svc.ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			Email:       "test@example.com",
			Code:        "123456",
			NewPassword: testPassword,
		})

		assertCode(t, err, codes.ResourceExhausted)
	})

	t.Run("wrong_code", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByEmail: func(_ context.Context, _ entities.GetUserByEmailDTO) (*entities.UserGetByEmail, Error.CodeError) {
				return &entities.UserGetByEmail{UserUUID: testUUID1, IsVerified: true}, ok()
			},
		}
		recRepo := &mockRecoveryRepo{
			getRecoveryCode: func(_ context.Context, _ entities.GetRecoveryCodeDTO) (string, Error.CodeError) {
				return "999999", ok()
			},
			incrRecoveryAttempts: func(_ context.Context, _ entities.IncrRecoveryAttemptsDTO) (int64, Error.CodeError) {
				return 1, ok()
			},
		}
		svc := buildSvc(svcDeps{user: userRepo, recovery: recRepo})

		_, err := svc.ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			Email:       "test@example.com",
			Code:        "123456",
			NewPassword: testPassword,
		})

		assertCode(t, err, codes.InvalidArgument)
	})
}

// ─── Verify2FA ────────────────────────────────────────────────────────────────

func TestVerify2FA(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		twoFARepo := &mockTwoFARepo{
			get2FAData: func(_ context.Context, _ entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError) {
				return &entities.TwoFAData{UserUUID: testUUID1, Code: "123456"}, ok()
			},
			incr2FAAttempts: func(_ context.Context, _ entities.Incr2FAAttemptsDTO) (int64, Error.CodeError) {
				return 1, ok()
			},
			delete2FAData: func(_ context.Context, _ entities.Delete2FADataDTO) Error.CodeError { return ok() },
		}
		authRepo := &mockAuthRepo{
			saveRefreshToken: func(_ context.Context, _ entities.SaveRefreshTokenDTO) Error.CodeError { return ok() },
		}
		svc := buildSvc(svcDeps{auth: authRepo, twoFA: twoFARepo})

		resp, err := svc.Verify2FA(context.Background(), &pb.Verify2FARequest{
			SessionUuid: testUUID1,
			Code:        "123456",
		})

		assertNoError(t, err)
		if resp.GetUserUuid() != testUUID1 {
			t.Errorf("expected user UUID %q, got %q", testUUID1, resp.GetUserUuid())
		}
		if resp.GetAccessToken() == "" || resp.GetRefreshToken() == "" {
			t.Error("expected non-empty tokens")
		}
	})

	t.Run("invalid_session_uuid", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.Verify2FA(context.Background(), &pb.Verify2FARequest{
			SessionUuid: "not-a-uuid",
			Code:        "123456",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid_code_format", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.Verify2FA(context.Background(), &pb.Verify2FARequest{
			SessionUuid: testUUID1,
			Code:        "abc",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("session_not_found", func(t *testing.T) {
		twoFARepo := &mockTwoFARepo{
			get2FAData: func(_ context.Context, _ entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError) {
				return nil, Error.Public(codes.NotFound, "2FA code not found")
			},
		}
		svc := buildSvc(svcDeps{twoFA: twoFARepo})

		_, err := svc.Verify2FA(context.Background(), &pb.Verify2FARequest{
			SessionUuid: testUUID1,
			Code:        "123456",
		})

		assertCode(t, err, codes.NotFound)
	})

	t.Run("too_many_attempts", func(t *testing.T) {
		twoFARepo := &mockTwoFARepo{
			get2FAData: func(_ context.Context, _ entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError) {
				return &entities.TwoFAData{UserUUID: testUUID1, Code: "123456"}, ok()
			},
			incr2FAAttempts: func(_ context.Context, _ entities.Incr2FAAttemptsDTO) (int64, Error.CodeError) {
				return max2FAAttempts + 1, ok()
			},
			delete2FAData: func(_ context.Context, _ entities.Delete2FADataDTO) Error.CodeError { return ok() },
		}
		svc := buildSvc(svcDeps{twoFA: twoFARepo})

		_, err := svc.Verify2FA(context.Background(), &pb.Verify2FARequest{
			SessionUuid: testUUID1,
			Code:        "123456",
		})

		assertCode(t, err, codes.ResourceExhausted)
	})

	t.Run("wrong_code", func(t *testing.T) {
		twoFARepo := &mockTwoFARepo{
			get2FAData: func(_ context.Context, _ entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError) {
				return &entities.TwoFAData{UserUUID: testUUID1, Code: "999999"}, ok()
			},
			incr2FAAttempts: func(_ context.Context, _ entities.Incr2FAAttemptsDTO) (int64, Error.CodeError) {
				return 1, ok()
			},
		}
		svc := buildSvc(svcDeps{twoFA: twoFARepo})

		_, err := svc.Verify2FA(context.Background(), &pb.Verify2FARequest{
			SessionUuid: testUUID1,
			Code:        "123456",
		})

		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("save_token_error", func(t *testing.T) {
		twoFARepo := &mockTwoFARepo{
			get2FAData: func(_ context.Context, _ entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError) {
				return &entities.TwoFAData{UserUUID: testUUID1, Code: "123456"}, ok()
			},
			incr2FAAttempts: func(_ context.Context, _ entities.Incr2FAAttemptsDTO) (int64, Error.CodeError) {
				return 1, ok()
			},
			delete2FAData: func(_ context.Context, _ entities.Delete2FADataDTO) Error.CodeError { return ok() },
		}
		authRepo := &mockAuthRepo{
			saveRefreshToken: func(_ context.Context, _ entities.SaveRefreshTokenDTO) Error.CodeError {
				return Error.Internal(fmt.Errorf("redis error"))
			},
		}
		svc := buildSvc(svcDeps{auth: authRepo, twoFA: twoFARepo})

		_, err := svc.Verify2FA(context.Background(), &pb.Verify2FARequest{
			SessionUuid: testUUID1,
			Code:        "123456",
		})

		assertCode(t, err, codes.Internal)
	})
}

// ─── UpdateUser2FA ────────────────────────────────────────────────────────────

func TestUpdateUser2FA(t *testing.T) {
	t.Run("enable_success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateUser2FA: func(_ context.Context, dto entities.UpdateUser2FADTO) Error.CodeError {
				if !dto.TwoFAEnabled {
					return Error.Internal(fmt.Errorf("expected enabled=true"))
				}
				return ok()
			},
		}
		svc := buildSvc(svcDeps{user: userRepo})

		_, err := svc.UpdateUser2FA(context.Background(), &pb.UpdateUser2FARequest{
			UserUuid:   testUUID1,
			Enable_2Fa: true,
		})

		assertNoError(t, err)
	})

	t.Run("disable_success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateUser2FA: func(_ context.Context, dto entities.UpdateUser2FADTO) Error.CodeError {
				if dto.TwoFAEnabled {
					return Error.Internal(fmt.Errorf("expected enabled=false"))
				}
				return ok()
			},
		}
		svc := buildSvc(svcDeps{user: userRepo})

		_, err := svc.UpdateUser2FA(context.Background(), &pb.UpdateUser2FARequest{
			UserUuid:   testUUID1,
			Enable_2Fa: false,
		})

		assertNoError(t, err)
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.UpdateUser2FA(context.Background(), &pb.UpdateUser2FARequest{
			UserUuid:   "not-a-uuid",
			Enable_2Fa: true,
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("user_not_found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateUser2FA: func(_ context.Context, _ entities.UpdateUser2FADTO) Error.CodeError {
				return Error.Public(codes.NotFound, "user not found")
			},
		}
		svc := buildSvc(svcDeps{user: userRepo})

		_, err := svc.UpdateUser2FA(context.Background(), &pb.UpdateUser2FARequest{
			UserUuid:   testUUID1,
			Enable_2Fa: true,
		})

		assertCode(t, err, codes.NotFound)
	})
}

// ─── GetVerificationCode ──────────────────────────────────────────────────────

func TestGetVerificationCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		verRepo := &mockVerificationRepo{
			getVerificationCode: func(_ context.Context, dto entities.GetVerificationCodeDTO) (string, Error.CodeError) {
				if dto.UserUUID != testUUID1 {
					return "", Error.Internal(fmt.Errorf("unexpected uuid"))
				}
				return "123456", ok()
			},
		}
		svc := buildSvc(svcDeps{verification: verRepo})

		resp, err := svc.GetVerificationCode(context.Background(), &pb.GetVerificationCodeRequest{
			UserUuid: testUUID1,
		})

		assertNoError(t, err)
		if resp.GetCode() != "123456" {
			t.Errorf("expected code '123456', got %q", resp.GetCode())
		}
	})

	t.Run("unavailable_in_production", func(t *testing.T) {
		svc := buildSvc(svcDeps{appEnv: "production"})

		_, err := svc.GetVerificationCode(context.Background(), &pb.GetVerificationCodeRequest{
			UserUuid: testUUID1,
		})

		assertCode(t, err, codes.Unimplemented)
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.GetVerificationCode(context.Background(), &pb.GetVerificationCodeRequest{
			UserUuid: "not-a-uuid",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("code_not_found", func(t *testing.T) {
		verRepo := &mockVerificationRepo{
			getVerificationCode: func(_ context.Context, _ entities.GetVerificationCodeDTO) (string, Error.CodeError) {
				return "", Error.Public(codes.NotFound, "code not found")
			},
		}
		svc := buildSvc(svcDeps{verification: verRepo})

		_, err := svc.GetVerificationCode(context.Background(), &pb.GetVerificationCodeRequest{
			UserUuid: testUUID1,
		})

		assertCode(t, err, codes.NotFound)
	})
}

// ─── GetRecoveryCode ──────────────────────────────────────────────────────────

func TestGetRecoveryCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		recRepo := &mockRecoveryRepo{
			getRecoveryCode: func(_ context.Context, dto entities.GetRecoveryCodeDTO) (string, Error.CodeError) {
				if dto.UserUUID != testUUID1 {
					return "", Error.Internal(fmt.Errorf("unexpected uuid"))
				}
				return "654321", ok()
			},
		}
		svc := buildSvc(svcDeps{recovery: recRepo})

		resp, err := svc.GetRecoveryCode(context.Background(), &pb.GetRecoveryCodeRequest{
			UserUuid: testUUID1,
		})

		assertNoError(t, err)
		if resp.GetCode() != "654321" {
			t.Errorf("expected code '654321', got %q", resp.GetCode())
		}
	})

	t.Run("unavailable_in_production", func(t *testing.T) {
		svc := buildSvc(svcDeps{appEnv: "production"})

		_, err := svc.GetRecoveryCode(context.Background(), &pb.GetRecoveryCodeRequest{
			UserUuid: testUUID1,
		})

		assertCode(t, err, codes.Unimplemented)
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.GetRecoveryCode(context.Background(), &pb.GetRecoveryCodeRequest{
			UserUuid: "not-a-uuid",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("code_not_found", func(t *testing.T) {
		recRepo := &mockRecoveryRepo{
			getRecoveryCode: func(_ context.Context, _ entities.GetRecoveryCodeDTO) (string, Error.CodeError) {
				return "", Error.Public(codes.NotFound, "code not found")
			},
		}
		svc := buildSvc(svcDeps{recovery: recRepo})

		_, err := svc.GetRecoveryCode(context.Background(), &pb.GetRecoveryCodeRequest{
			UserUuid: testUUID1,
		})

		assertCode(t, err, codes.NotFound)
	})
}

// ─── Get2FACode ───────────────────────────────────────────────────────────────

func TestGet2FACode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		twoFARepo := &mockTwoFARepo{
			get2FAData: func(_ context.Context, dto entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError) {
				if dto.SessionUUID != testUUID1 {
					return nil, Error.Internal(fmt.Errorf("unexpected session uuid"))
				}
				return &entities.TwoFAData{UserUUID: testUUID2, Code: "111222"}, ok()
			},
		}
		svc := buildSvc(svcDeps{twoFA: twoFARepo})

		resp, err := svc.Get2FACode(context.Background(), &pb.Get2FACodeRequest{
			SessionUuid: testUUID1,
		})

		assertNoError(t, err)
		if resp.GetCode() != "111222" {
			t.Errorf("expected code '111222', got %q", resp.GetCode())
		}
	})

	t.Run("unavailable_in_production", func(t *testing.T) {
		svc := buildSvc(svcDeps{appEnv: "production"})

		_, err := svc.Get2FACode(context.Background(), &pb.Get2FACodeRequest{
			SessionUuid: testUUID1,
		})

		assertCode(t, err, codes.Unimplemented)
	})

	t.Run("invalid_session_uuid", func(t *testing.T) {
		svc := buildSvc(svcDeps{})

		_, err := svc.Get2FACode(context.Background(), &pb.Get2FACodeRequest{
			SessionUuid: "not-a-uuid",
		})

		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("data_not_found", func(t *testing.T) {
		twoFARepo := &mockTwoFARepo{
			get2FAData: func(_ context.Context, _ entities.Get2FADataDTO) (*entities.TwoFAData, Error.CodeError) {
				return nil, Error.Public(codes.NotFound, "2FA code not found")
			},
		}
		svc := buildSvc(svcDeps{twoFA: twoFARepo})

		_, err := svc.Get2FACode(context.Background(), &pb.Get2FACodeRequest{
			SessionUuid: testUUID1,
		})

		assertCode(t, err, codes.NotFound)
	})
}
