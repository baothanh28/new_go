package auth

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"myapp/internal/pkg/database"
)

// Repository provides database operations for users
type Repository struct {
	*database.MasterRepo[User]
}

// NewRepository creates a new user repository using master database
func NewRepository(dbManager *database.DatabaseManager) *Repository {
	// Use master database for user authentication
	return &Repository{
		MasterRepo: database.NewMasterRepo[User](dbManager),
	}
}

// GetByEmail retrieves a user by email address
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.GetDB().WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ErrUserNotFound{Email: email}
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

// EmailExists checks if a user with the given email already exists
func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check email exists: %w", err)
	}
	return count > 0, nil
}

// GetByID retrieves a user by ID
func (r *Repository) GetByID(ctx context.Context, id uint) (*User, error) {
	user, err := r.MasterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// Create creates a new user
func (r *Repository) Create(ctx context.Context, user *User) error {
	return r.MasterRepo.Insert(ctx, user)
}
