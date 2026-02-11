// +build cgo

package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"myapp/internal/pkg/auth"
	"myapp/internal/pkg/database"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate auth tables
	err = db.AutoMigrate(&auth.User{}, &auth.RefreshToken{}, &auth.TokenBlacklist{})
	require.NoError(t, err)

	return db
}

// setupTestRepository creates a test repository
func setupTestRepository(t *testing.T) (*auth.Repository, *gorm.DB) {
	db := setupTestDB(t)
	
	// Create a mock DatabaseManager
	dbManager := &database.DatabaseManager{
		MasterDB: db,
	}
	
	repo := auth.NewRepository(dbManager)
	return repo, db
}

func TestRepository_GetByEmail(t *testing.T) {
	repo, db := setupTestRepository(t)
	ctx := context.Background()

	// Create test user
	user := &User{
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "user",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	t.Run("existing user", func(t *testing.T) {
		found, err := repo.GetByEmail(ctx, "test@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.Email, found.Email)
		assert.Equal(t, user.Role, found.Role)
	})

	t.Run("non-existent user", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
		assert.IsType(t, &auth.ErrUserNotFound{}, err)
	})
}

func TestRepository_EmailExists(t *testing.T) {
	repo, db := setupTestRepository(t)
	ctx := context.Background()

	// Create test user
	user := &User{
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "user",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	t.Run("existing email", func(t *testing.T) {
		exists, err := repo.EmailExists(ctx, "test@example.com")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existent email", func(t *testing.T) {
		exists, err := repo.EmailExists(ctx, "nonexistent@example.com")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestRepository_GetByID(t *testing.T) {
	repo, db := setupTestRepository(t)
	ctx := context.Background()

	// Create test user
	user := &User{
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "user",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	t.Run("existing user", func(t *testing.T) {
		found, err := repo.GetByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("non-existent user", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 999)
		assert.Error(t, err)
	})
}

func TestRepository_Create(t *testing.T) {
	repo, _ := setupTestRepository(t)
	ctx := context.Background()

	t.Run("create new user", func(t *testing.T) {
		user := &auth.User{
			Email:    "newuser@example.com",
			Password: "hashed_password",
			Role:     "user",
		}

		err := repo.Create(ctx, user)
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.NotZero(t, user.CreatedAt)
	})

	t.Run("create user with duplicate email", func(t *testing.T) {
		user1 := &User{
			Email:    "duplicate@example.com",
			Password: "hashed_password",
			Role:     "user",
		}
		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		user2 := &User{
			Email:    "duplicate@example.com",
			Password: "hashed_password",
			Role:     "user",
		}
		err = repo.Create(ctx, user2)
		assert.Error(t, err) // Should fail due to unique constraint
	})
}
