package services

import (
	"backend/internal/entities"
	"backend/internal/repositories"
	"context"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, userRequest *entities.CreateUserRequest) (*entities.CreateUserResponse, error) {
	user := &entities.User{}

	user.Login = userRequest.Login
	user.FirstName = userRequest.FirstName
	user.SecondName = userRequest.SecondName
	user.ThirdName = userRequest.ThirdName
	user.Role = userRequest.Role
	user.Email = ""

	userPassword := "abcabcabc"
	//Password create logic
	//...
	user.PasswordHash = userPassword
	user.PasswordSalt = userPassword

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	createUserResponse := &entities.CreateUserResponse{
		ID:       user.ID,
		Login:    user.Login,
		Password: user.PasswordHash,
	}

	return createUserResponse, nil
}
