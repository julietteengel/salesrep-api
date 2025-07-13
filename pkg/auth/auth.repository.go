package auth

import (
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IAuthRepository interface {
	GetUserAuthData(ctx context.Context, userAuth0ID string) (AuthData, error)
}

type AuthRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(database *gorm.DB) IAuthRepository {
	return &AuthRepository{db: database}
}

func (repository *AuthRepository) GetUserAuthData(ctx context.Context, userAuth0ID string) (AuthData, error) {
	var authData AuthData

	// Get user data
	result := repository.db.WithContext(ctx).
		Table("logins").
		Joins("INNER JOIN users ON users.id = logins.user_id").
		Where("logins.auth0_id = ?", userAuth0ID).
		Where("logins.deleted_at IS NULL").
		Where("users.deleted_at IS NULL").
		Select("users.id as user_id, users.company_id, users.status").
		First(&authData.BaseAuthData)
	if result.Error != nil {
		return AuthData{}, errors.WithStack(result.Error)
	}


	return authData, nil
}
