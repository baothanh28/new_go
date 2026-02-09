package auth

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"myapp/internal/pkg/database"
)

// Repository provides database operations for users
type Repository struct {
	*database.BaseRepository[User]
}

// NewRepository creates a new user repository
func NewRepository(dbManager *database.DatabaseManager) *Repository {
	// Use master database for user authentication
	return &Repository{
		BaseRepository: database.NewBaseRepository[User](dbManager.MasterDB),
	}
}

// GetByEmail retrieves a user by email address
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.GetDB().WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

// GetByRole retrieves all users with a specific role
func (r *Repository) GetByRole(ctx context.Context, role string) ([]*User, error) {
	var users []*User
	if err := r.GetDB().WithContext(ctx).Where("role = ?", role).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("get users by role %s: %w", role, err)
	}
	return users, nil
}

// EmailExists checks if a user with the given email already exists
func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check email exists: %w", err)
	}
	return count > 0, nil
}
