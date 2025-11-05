package auth

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"microapi/internal/grpc/auth/entities"
	"microapi/protos/auth_service"
)

type Auth interface {
	Login(ctx context.Context, req *auth_entities.LoginRequest) (*auth_entities.LoginResponse, error)
	Register(ctx context.Context, req *auth_entities.RegisterRequest) (*auth_entities.RegisterResponse, error)
	GetUser(ctx context.Context, req *auth_entities.GetUserRequest) (*auth_entities.GetUserResponse, error)
	RefreshToken(ctx context.Context, req *auth_entities.RefreshTokenRequest) (*auth_entities.RefreshTokenResponse, error)
	UpdateUser(ctx context.Context, req *auth_entities.UpdateUserRequest) (*auth_entities.UpdateUserResponse, error)
	RevokeToken(ctx context.Context, req *auth_entities.RevokeTokenRequest) error
	DeleteUser(ctx context.Context, req *auth_entities.DeleteUserRequest) error
}

type ServerAPI struct {
	auth_proto.UnimplementedAuthServiceServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	auth_proto.RegisterAuthServiceServer(gRPC, &ServerAPI{auth: auth})
}

func (s *ServerAPI) Register(ctx context.Context, req *auth_proto.RegisterRequest) (*auth_proto.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	return &auth_proto.RegisterResponse{
		UserId:       1,
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
	}, nil
}

func (s *ServerAPI) Login(ctx context.Context, req *auth_proto.LoginRequest) (*auth_proto.LoginResponse, error) {
	panic("implement me")
}

func (s *ServerAPI) GetUser(ctx context.Context, req *auth_proto.GetUserRequest) (*auth_proto.GetUserResponse, error) {
	panic("implement me")
}

func (s *ServerAPI) RefreshToken(ctx context.Context, req *auth_proto.RefreshTokenRequest) (*auth_proto.RefreshTokenResponse, error) {
	panic("implement me")
}

func (s *ServerAPI) UpdateUser(ctx context.Context, req *auth_proto.UpdateUserRequest) (*auth_proto.UpdateUserResponse, error) {
	panic("implement me")
}

func (s *ServerAPI) RevokeToken(ctx context.Context, req *auth_proto.RevokeTokenRequest) (*emptypb.Empty, error) {
	panic("implement me")
}

func (s *ServerAPI) DeleteUser(ctx context.Context, req *auth_proto.DeleteUserRequest) (*emptypb.Empty, error) {
	panic("implement me")
}

// *********************
// Validators
// *********************

func validateRegister(req *auth_proto.RegisterRequest) error {
	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "login is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetFirstName() == "" {
		return status.Error(codes.InvalidArgument, "first name is required")
	}

	if req.GetSecondName() == "" {
		return status.Error(codes.InvalidArgument, "second name is required")
	}

	return nil
}
