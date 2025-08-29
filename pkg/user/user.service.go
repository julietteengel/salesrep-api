package user

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
)

// Response schema for GetOrCreateUser
type UserResponse struct {
	ID            uint   `json:"id"`
	Auth0ID       string `json:"auth0_id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Role          string `json:"role"`
	Language      string `json:"language"`
	Name          string `json:"name,omitempty"`
	Avatar        string `json:"avatar,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type IUserService interface {
	GetOrCreateUser(ctx context.Context, auth0ID string, email string) (*UserResponse, error)
	GetByAuth0ID(ctx context.Context, auth0ID string) (*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
	UpdateEmailVerificationStatus(ctx context.Context, auth0ID string, verified bool) error
}

type UserService struct {
	repo IUserRepository
}

func NewUserService(repo IUserRepository) IUserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetOrCreateUser(ctx context.Context, auth0ID string, email string) (*UserResponse, error) {
	// Try to find existing user
	spew.Dump("kikoo")
	existingUser, err := s.repo.FindByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	var user *User

	if existingUser != nil {
		// User exists, update email if changed
		if existingUser.Email != email {
			existingUser.Email = email
			existingUser.EmailVerified = true // Assuming email is verified if coming from Auth0
			if err := s.repo.Update(ctx, existingUser); err != nil {
				return nil, fmt.Errorf("failed to update user: %w", err)
			}
		}
		user = existingUser
	} else {
		// Create new user
		newUser := &User{
			Auth0ID:       auth0ID,
			Email:         email,
			EmailVerified: true, // Assuming email is verified if coming from Auth0
			Role:          SalesRepRole,
			Language:      "en",
		}

		if err := s.repo.Create(ctx, newUser); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		user = newUser
	}

	// Convert to response format
	response := &UserResponse{
		ID:            user.ID,
		Auth0ID:       user.Auth0ID,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Role:          string(user.Role),
		Language:      user.Language,
		Name:          user.GetFullName(),
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if user.Avatar != nil {
		response.Avatar = *user.Avatar
	}

	return response, nil
}

func (s *UserService) GetByAuth0ID(ctx context.Context, auth0ID string) (*User, error) {
	user, err := s.repo.FindByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *UserService) UpdateEmailVerificationStatus(ctx context.Context, auth0ID string, verified bool) error {
	return s.repo.UpdateEmailVerificationStatus(ctx, auth0ID, verified)
}
