package handlers

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	auth_proto "github.com/unwelcome/FrameWorkTask1/backend/auth/api/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/entities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ─── Register ─────────────────────────────────────────────────────────────────

func TestRegister(t *testing.T) {
	app := newApp(http.MethodPost, "/register",
		newAuthHandler(&mockAuthClient{
			register: func(_ context.Context, _ *auth_proto.RegisterRequest, _ ...grpc.CallOption) (*auth_proto.RegisterResponse, error) {
				return &auth_proto.RegisterResponse{UserUuid: userID}, nil
			},
		}).Register,
	)

	t.Run("success", func(t *testing.T) {
		body := fmt.Sprintf(`{"email":"test@example.com","password":"Password123","first_name":"Ivan","last_name":"Petrov","patronymic":"Ivanovich"}`)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/register", body))
		assertStatus(t, resp, fiber.StatusCreated)

		res := decodeBody[entities.RegisterResponse](t, resp)
		if res.UserUUID == "" {
			t.Error("expected non-empty user_uuid")
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		body := `{"email":"not-an-email","password":"Password123","first_name":"Ivan","last_name":"Petrov","patronymic":"Ivanovich"}`
		resp, _ := app.Test(jsonReq(http.MethodPost, "/register", body))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("missing password", func(t *testing.T) {
		body := `{"email":"test@example.com","first_name":"Ivan","last_name":"Petrov","patronymic":"Ivanovich"}`
		resp, _ := app.Test(jsonReq(http.MethodPost, "/register", body))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - conflict (email taken)", func(t *testing.T) {
		appConflict := newApp(http.MethodPost, "/register",
			newAuthHandler(&mockAuthClient{
				register: func(_ context.Context, _ *auth_proto.RegisterRequest, _ ...grpc.CallOption) (*auth_proto.RegisterResponse, error) {
					return nil, grpcErr(codes.AlreadyExists)
				},
			}).Register,
		)
		body := `{"email":"test@example.com","password":"Password123","first_name":"Ivan","last_name":"Petrov","patronymic":"Ivanovich"}`
		resp, _ := appConflict.Test(jsonReq(http.MethodPost, "/register", body))
		assertStatus(t, resp, fiber.StatusConflict)
	})

	t.Run("service internal error", func(t *testing.T) {
		appErr := newApp(http.MethodPost, "/register",
			newAuthHandler(&mockAuthClient{
				register: func(_ context.Context, _ *auth_proto.RegisterRequest, _ ...grpc.CallOption) (*auth_proto.RegisterResponse, error) {
					return nil, grpcErr(codes.Internal)
				},
			}).Register,
		)
		body := `{"email":"test@example.com","password":"Password123","first_name":"Ivan","last_name":"Petrov","patronymic":"Ivanovich"}`
		resp, _ := appErr.Test(jsonReq(http.MethodPost, "/register", body))
		assertStatus(t, resp, fiber.StatusInternalServerError)
	})
}

// ─── Login ────────────────────────────────────────────────────────────────────

func TestLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPost, "/login",
			newAuthHandler(&mockAuthClient{
				login: func(_ context.Context, _ *auth_proto.LoginRequest, _ ...grpc.CallOption) (*auth_proto.LoginResponse, error) {
					return &auth_proto.LoginResponse{
						UserUuid:     userID,
						AccessToken:  "access.token.here",
						RefreshToken: "refresh.token.here",
					}, nil
				},
			}).Login,
		)
		body := `{"email":"test@example.com","password":"Password123"}`
		resp, _ := app.Test(jsonReq(http.MethodPost, "/login", body))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.LoginResponse](t, resp)
		if res.UserUUID != userID {
			t.Errorf("expected user_uuid=%q, got %q", userID, res.UserUUID)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		app := newApp(http.MethodPost, "/login",
			newAuthHandler(&mockAuthClient{}).Login,
		)
		body := `{"email":"bad","password":"Password123"}`
		resp, _ := app.Test(jsonReq(http.MethodPost, "/login", body))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodPost, "/login",
			newAuthHandler(&mockAuthClient{
				login: func(_ context.Context, _ *auth_proto.LoginRequest, _ ...grpc.CallOption) (*auth_proto.LoginResponse, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).Login,
		)
		body := `{"email":"test@example.com","password":"Password123"}`
		resp, _ := app.Test(jsonReq(http.MethodPost, "/login", body))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}

// ─── GetUser ──────────────────────────────────────────────────────────────────

func TestGetUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodGet, "/user/:user_uuid/info",
			newAuthHandler(&mockAuthClient{
				getUser: func(_ context.Context, in *auth_proto.GetUserRequest, _ ...grpc.CallOption) (*auth_proto.GetUserResponse, error) {
					return &auth_proto.GetUserResponse{
						UserUuid:  in.GetUserUuid(),
						Email:     "test@example.com",
						FirstName: "Ivan",
						LastName:  "Petrov",
					}, nil
				},
			}).GetUser,
		)
		resp, _ := app.Test(getReq("/user/" + userID + "/info"))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.GetUserResponse](t, resp)
		if res.UserUUID != userID {
			t.Errorf("expected user_uuid=%q, got %q", userID, res.UserUUID)
		}
		if res.Email != "test@example.com" {
			t.Errorf("expected email, got %q", res.Email)
		}
	})

	t.Run("invalid uuid in path", func(t *testing.T) {
		app := newApp(http.MethodGet, "/user/:user_uuid/info",
			newAuthHandler(&mockAuthClient{}).GetUser,
		)
		resp, _ := app.Test(getReq("/user/not-a-uuid/info"))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodGet, "/user/:user_uuid/info",
			newAuthHandler(&mockAuthClient{
				getUser: func(_ context.Context, _ *auth_proto.GetUserRequest, _ ...grpc.CallOption) (*auth_proto.GetUserResponse, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).GetUser,
		)
		resp, _ := app.Test(getReq("/user/" + userID + "/info"))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}

// ─── ChangePassword ───────────────────────────────────────────────────────────

func TestChangePassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/user/password",
			newAuthHandler(&mockAuthClient{
				changePassword: func(_ context.Context, _ *auth_proto.ChangePasswordRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).ChangePassword,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/user/password", `{"password":"NewPass123"}`))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("password too short", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/user/password",
			newAuthHandler(&mockAuthClient{}).ChangePassword,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/user/password", `{"password":"Ab1"}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/user/password",
			newAuthHandler(&mockAuthClient{
				changePassword: func(_ context.Context, _ *auth_proto.ChangePasswordRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.Internal)
				},
			}).ChangePassword,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/user/password", `{"password":"NewPass123"}`))
		assertStatus(t, resp, fiber.StatusInternalServerError)
	})
}

// ─── UpdateUserBio ────────────────────────────────────────────────────────────

func TestUpdateUserBio(t *testing.T) {
	validBody := `{"first_name":"Ivan","last_name":"Petrov","patronymic":"Ivanovich"}`

	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/user/bio",
			newAuthHandler(&mockAuthClient{
				updateUserBio: func(_ context.Context, _ *auth_proto.UpdateUserBioRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).UpdateUserBio,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/user/bio", validBody))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("missing first_name", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/user/bio",
			newAuthHandler(&mockAuthClient{}).UpdateUserBio,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/user/bio", `{"last_name":"Petrov","patronymic":"Ivanovich"}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodPatch, "/user/bio",
			newAuthHandler(&mockAuthClient{
				updateUserBio: func(_ context.Context, _ *auth_proto.UpdateUserBioRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).UpdateUserBio,
		)
		resp, _ := app.Test(jsonReq(http.MethodPatch, "/user/bio", validBody))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}

// ─── DeleteUser ───────────────────────────────────────────────────────────────

func TestDeleteUser(t *testing.T) {
	validBody := fmt.Sprintf(`{"target_uuid":%q}`, targetID)

	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/user/account",
			newAuthHandler(&mockAuthClient{
				deleteUser: func(_ context.Context, _ *auth_proto.DeleteUserRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).DeleteUser,
		)
		resp, _ := app.Test(jsonReq(http.MethodDelete, "/user/account", validBody))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("missing target_uuid", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/user/account",
			newAuthHandler(&mockAuthClient{}).DeleteUser,
		)
		resp, _ := app.Test(jsonReq(http.MethodDelete, "/user/account", `{}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/user/account",
			newAuthHandler(&mockAuthClient{
				deleteUser: func(_ context.Context, _ *auth_proto.DeleteUserRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).DeleteUser,
		)
		resp, _ := app.Test(jsonReq(http.MethodDelete, "/user/account", validBody))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}

// ─── GetAllActiveTokens ───────────────────────────────────────────────────────

func TestGetAllActiveTokens(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodGet, "/user/tokens",
			newAuthHandler(&mockAuthClient{
				getAllActiveTokens: func(_ context.Context, _ *auth_proto.GetAllActiveTokensRequest, _ ...grpc.CallOption) (*auth_proto.GetAllActiveTokensResponse, error) {
					return &auth_proto.GetAllActiveTokensResponse{
						Tokens: []*auth_proto.Token{
							{Token: validJWT},
						},
					}, nil
				},
			}).GetAllActiveTokens,
		)
		resp, _ := app.Test(getReq("/user/tokens"))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.GetAllActiveTokensResponse](t, resp)
		if len(res.Tokens) != 1 {
			t.Errorf("expected 1 token, got %d", len(res.Tokens))
		}
	})

	t.Run("service error", func(t *testing.T) {
		app := newApp(http.MethodGet, "/user/tokens",
			newAuthHandler(&mockAuthClient{
				getAllActiveTokens: func(_ context.Context, _ *auth_proto.GetAllActiveTokensRequest, _ ...grpc.CallOption) (*auth_proto.GetAllActiveTokensResponse, error) {
					return nil, grpcErr(codes.Internal)
				},
			}).GetAllActiveTokens,
		)
		resp, _ := app.Test(getReq("/user/tokens"))
		assertStatus(t, resp, fiber.StatusInternalServerError)
	})
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

func TestRefreshToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodPost, "/refresh",
			newAuthHandler(&mockAuthClient{
				refreshToken: func(_ context.Context, _ *auth_proto.RefreshTokenRequest, _ ...grpc.CallOption) (*auth_proto.RefreshTokenResponse, error) {
					return &auth_proto.RefreshTokenResponse{
						AccessToken:  "new.access.token",
						RefreshToken: "new.refresh.token",
					}, nil
				},
			}).RefreshToken,
		)
		body := fmt.Sprintf(`{"refresh_token":%q}`, validJWT)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/refresh", body))
		assertStatus(t, resp, fiber.StatusOK)

		res := decodeBody[entities.RefreshTokenResponse](t, resp)
		if res.AccessToken == "" {
			t.Error("expected non-empty access_token")
		}
	})

	t.Run("invalid token format", func(t *testing.T) {
		app := newApp(http.MethodPost, "/refresh",
			newAuthHandler(&mockAuthClient{}).RefreshToken,
		)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/refresh", `{"refresh_token":"bad"}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - token expired", func(t *testing.T) {
		app := newApp(http.MethodPost, "/refresh",
			newAuthHandler(&mockAuthClient{
				refreshToken: func(_ context.Context, _ *auth_proto.RefreshTokenRequest, _ ...grpc.CallOption) (*auth_proto.RefreshTokenResponse, error) {
					return nil, grpcErr(codes.Unauthenticated)
				},
			}).RefreshToken,
		)
		body := fmt.Sprintf(`{"refresh_token":%q}`, validJWT)
		resp, _ := app.Test(jsonReq(http.MethodPost, "/refresh", body))
		assertStatus(t, resp, fiber.StatusUnauthorized)
	})
}

// ─── RevokeToken ──────────────────────────────────────────────────────────────

func TestRevokeToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/user/revoke/token",
			newAuthHandler(&mockAuthClient{
				revokeToken: func(_ context.Context, _ *auth_proto.RevokeTokenRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).RevokeToken,
		)
		body := fmt.Sprintf(`{"refresh_token":%q}`, validJWT)
		resp, _ := app.Test(jsonReq(http.MethodDelete, "/user/revoke/token", body))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("invalid token format", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/user/revoke/token",
			newAuthHandler(&mockAuthClient{}).RevokeToken,
		)
		resp, _ := app.Test(jsonReq(http.MethodDelete, "/user/revoke/token", `{"refresh_token":""}`))
		assertStatus(t, resp, fiber.StatusBadRequest)
	})

	t.Run("service error - not found", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/user/revoke/token",
			newAuthHandler(&mockAuthClient{
				revokeToken: func(_ context.Context, _ *auth_proto.RevokeTokenRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.NotFound)
				},
			}).RevokeToken,
		)
		body := fmt.Sprintf(`{"refresh_token":%q}`, validJWT)
		resp, _ := app.Test(jsonReq(http.MethodDelete, "/user/revoke/token", body))
		assertStatus(t, resp, fiber.StatusNotFound)
	})
}

// ─── RevokeAllTokens ──────────────────────────────────────────────────────────

func TestRevokeAllTokens(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/user/revoke/all",
			newAuthHandler(&mockAuthClient{
				revokeAllTokens: func(_ context.Context, _ *auth_proto.RevokeAllTokensRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				},
			}).RevokeAllTokens,
		)
		resp, _ := app.Test(deleteReq("/user/revoke/all"))
		assertStatus(t, resp, fiber.StatusOK)
	})

	t.Run("service error", func(t *testing.T) {
		app := newApp(http.MethodDelete, "/user/revoke/all",
			newAuthHandler(&mockAuthClient{
				revokeAllTokens: func(_ context.Context, _ *auth_proto.RevokeAllTokensRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, grpcErr(codes.Internal)
				},
			}).RevokeAllTokens,
		)
		resp, _ := app.Test(deleteReq("/user/revoke/all"))
		assertStatus(t, resp, fiber.StatusInternalServerError)
	})
}
