package user

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type IUserRepository interface {
	FindByAuth0ID(ctx context.Context, auth0ID string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	UpdateEmailVerificationStatus(ctx context.Context, auth0ID string, verified bool) error
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByAuth0ID(ctx context.Context, auth0ID string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("auth0_id = ?", auth0ID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil for not found
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) Update(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) UpdateEmailVerificationStatus(ctx context.Context, auth0ID string, verified bool) error {
	return r.db.WithContext(ctx).Model(&User{}).
		Where("auth0_id = ?", auth0ID).
		Update("email_verified", verified).Error
}