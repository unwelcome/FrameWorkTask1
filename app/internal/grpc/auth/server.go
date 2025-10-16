package auth

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"microapi/protos/auth_service"
)

type ServerAPI struct {
	auth_proto.UnimplementedAuthServiceServer
}

func Register(gRPC *grpc.Server) {
	auth_proto.RegisterAuthServiceServer(gRPC, &ServerAPI{})
}

func (s *ServerAPI) Register(ctx context.Context, req *auth_proto.RegisterRequest) (*auth_proto.RegisterResponse, error) {
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
