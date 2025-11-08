package handlers

import (
	pb "github.com/unwelcome/FrameWorkTask1/v1/gateway/api/auth"
)

type AuthHandler interface {
}

type authHandler struct {
	AuthServiceClient pb.AuthServiceClient
}

func NewAuthHandler(authServiceClient pb.AuthServiceClient) AuthHandler {
	return &authHandler{AuthServiceClient: authServiceClient}
}
