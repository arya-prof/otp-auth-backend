package service

import (
	"context"
	"fmt"

	"otp-auth-backend/models"
	"otp-auth-backend/store"
)

type UserService struct {
	userRepo *store.UserRepository
}

func NewUserService(userRepo *store.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *UserService) ListUsers(ctx context.Context, query *models.UserQuery) (*models.UserListResponse, error) {
	// Set default values
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	users, err := s.userRepo.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}
