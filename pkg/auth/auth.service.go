package auth

import "context"

type IAuthService interface {
	GetUserAuthData(ctx context.Context, userAuth0ID string) (AuthData, error)
}

type AuthService struct {
	authRepository IAuthRepository
}

func NewAuthService(authRepository IAuthRepository) IAuthService {
	return &AuthService{authRepository}
}

func (authService AuthService) GetUserAuthData(ctx context.Context, userAuth0ID string) (AuthData, error) {
	return authService.authRepository.GetUserAuthData(ctx, userAuth0ID)
}
