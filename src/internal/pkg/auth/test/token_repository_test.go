// +build cgo

package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"myapp/internal/pkg/auth"
	"myapp/internal/pkg/database"
)

// setupTestTokenRepository creates a test token repository
func setupTestTokenRepository(t *testing.T) (*auth.TokenRepository, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate auth tables
	err = db.AutoMigrate(&auth.User{}, &auth.RefreshToken{}, &auth.TokenBlacklist{})
	require.NoError(t, err)

	// Create a mock DatabaseManager
	dbManager := &database.DatabaseManager{
		MasterDB: db,
	}

	repo := auth.NewTokenRepository(dbManager)
	return repo, db
}

func TestTokenRepository_SaveRefreshToken(t *testing.T) {
	repo, db := setupTestTokenRepository(t)
	ctx := context.Background()

	// Create test user
		user := &auth.User{
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "user",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	t.Run("save refresh token", func(t *testing.T) {
		tokenHash := "hashed_refresh_token_123"
		expiresAt := time.Now().Add(7 * 24 * time.Hour)

		err := repo.SaveRefreshToken(ctx, user.ID, tokenHash, expiresAt)
		assert.NoError(t, err)

		// Verify token was saved
		var savedToken RefreshToken
		err = db.Where("token = ?", tokenHash).First(&savedToken).Error
		assert.NoError(t, err)
		assert.Equal(t, user.ID, savedToken.UserID)
		assert.Equal(t, tokenHash, savedToken.Token)
		assert.False(t, savedToken.Revoked)
	})
}

func TestTokenRepository_GetRefreshToken(t *testing.T) {
	repo, db := setupTestTokenRepository(t)
	ctx := context.Background()

	// Create test user
		user := &auth.User{
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "user",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	tokenHash := "hashed_refresh_token_123"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	err = repo.SaveRefreshToken(ctx, user.ID, tokenHash, expiresAt)
	require.NoError(t, err)

	t.Run("get existing token", func(t *testing.T) {
		token, err := repo.GetRefreshToken(ctx, tokenHash)
		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, user.ID, token.UserID)
		assert.Equal(t, tokenHash, token.Token)
		assert.False(t, token.Revoked)
	})

	t.Run("get non-existent token", func(t *testing.T) {
		_, err := repo.GetRefreshToken(ctx, "nonexistent_token")
		assert.Error(t, err)
		assert.IsType(t, &auth.ErrRefreshTokenNotFound{}, err)
	})

	t.Run("get revoked token", func(t *testing.T) {
		// Revoke the token
		err := repo.RevokeRefreshToken(ctx, tokenHash)
		require.NoError(t, err)

		// Should not find revoked token
		_, err = repo.GetRefreshToken(ctx, tokenHash)
		assert.Error(t, err)
	})

	t.Run("get expired token", func(t *testing.T) {
		expiredTokenHash := "expired_token_hash"
		expiredAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
		err := repo.SaveRefreshToken(ctx, user.ID, expiredTokenHash, expiredAt)
		require.NoError(t, err)

		_, err = repo.GetRefreshToken(ctx, expiredTokenHash)
		assert.Error(t, err)
		assert.IsType(t, &auth.ErrTokenExpired{}, err)
	})
}

func TestTokenRepository_RevokeRefreshToken(t *testing.T) {
	repo, db := setupTestTokenRepository(t)
	ctx := context.Background()

	// Create test user
		user := &auth.User{
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "user",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	tokenHash := "hashed_refresh_token_123"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	err = repo.SaveRefreshToken(ctx, user.ID, tokenHash, expiresAt)
	require.NoError(t, err)

	t.Run("revoke token", func(t *testing.T) {
		err := repo.RevokeRefreshToken(ctx, tokenHash)
		assert.NoError(t, err)

		// Verify token is revoked
		var token auth.RefreshToken
		err = db.Where("token = ?", tokenHash).First(&token).Error
		assert.NoError(t, err)
		assert.True(t, token.Revoked)
		assert.NotNil(t, token.RevokedAt)
	})
}

func TestTokenRepository_RevokeAllUserTokens(t *testing.T) {
	repo, db := setupTestTokenRepository(t)
	ctx := context.Background()

	// Create test user
		user := &auth.User{
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "user",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create multiple tokens for the user
	token1 := "token_hash_1"
	token2 := "token_hash_2"
	token3 := "token_hash_3"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	err = repo.SaveRefreshToken(ctx, user.ID, token1, expiresAt)
	require.NoError(t, err)
	err = repo.SaveRefreshToken(ctx, user.ID, token2, expiresAt)
	require.NoError(t, err)
	err = repo.SaveRefreshToken(ctx, user.ID, token3, expiresAt)
	require.NoError(t, err)

	t.Run("revoke all user tokens", func(t *testing.T) {
		err := repo.RevokeAllUserTokens(ctx, user.ID)
		assert.NoError(t, err)

		// Verify all tokens are revoked
		var tokens []auth.RefreshToken
		err = db.Where("user_id = ?", user.ID).Find(&tokens).Error
		assert.NoError(t, err)
		for _, token := range tokens {
			assert.True(t, token.Revoked)
		}
	})
}

func TestTokenRepository_AddToBlacklist(t *testing.T) {
	repo, _ := setupTestTokenRepository(t)
	ctx := context.Background()

	t.Run("add to blacklist", func(t *testing.T) {
		jti := "test_jti_123"
		expiresAt := time.Now().Add(15 * time.Minute)

		err := repo.AddToBlacklist(ctx, jti, expiresAt)
		assert.NoError(t, err)

		// Verify it's blacklisted
		isBlacklisted, err := repo.IsBlacklisted(ctx, jti)
		assert.NoError(t, err)
		assert.True(t, isBlacklisted)
	})
}

func TestTokenRepository_IsBlacklisted(t *testing.T) {
	repo, _ := setupTestTokenRepository(t)
	ctx := context.Background()

	jti := "test_jti_123"
	expiresAt := time.Now().Add(15 * time.Minute)

	err := repo.AddToBlacklist(ctx, jti, expiresAt)
	require.NoError(t, err)

	t.Run("blacklisted token", func(t *testing.T) {
		isBlacklisted, err := repo.IsBlacklisted(ctx, jti)
		assert.NoError(t, err)
		assert.True(t, isBlacklisted)
	})

	t.Run("non-blacklisted token", func(t *testing.T) {
		isBlacklisted, err := repo.IsBlacklisted(ctx, "non_blacklisted_jti")
		assert.NoError(t, err)
		assert.False(t, isBlacklisted)
	})
}

func TestTokenRepository_CleanupExpiredTokens(t *testing.T) {
	repo, db := setupTestTokenRepository(t)
	ctx := context.Background()

	// Create test user
		user := &auth.User{
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "user",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create expired refresh token
	expiredTokenHash := "expired_token"
	expiredAt := time.Now().Add(-1 * time.Hour)
	err = repo.SaveRefreshToken(ctx, user.ID, expiredTokenHash, expiredAt)
	require.NoError(t, err)

	// Create expired blacklist entry
	expiredJTI := "expired_jti"
	expiredBlacklistAt := time.Now().Add(-1 * time.Hour)
	err = repo.AddToBlacklist(ctx, expiredJTI, expiredBlacklistAt)
	require.NoError(t, err)

	// Create valid entries
	validTokenHash := "valid_token"
	validExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	err = repo.SaveRefreshToken(ctx, user.ID, validTokenHash, validExpiresAt)
	require.NoError(t, err)

	validJTI := "valid_jti"
	validBlacklistAt := time.Now().Add(15 * time.Minute)
	err = repo.AddToBlacklist(ctx, validJTI, validBlacklistAt)
	require.NoError(t, err)

	t.Run("cleanup expired tokens", func(t *testing.T) {
		err := repo.CleanupExpiredTokens(ctx)
		assert.NoError(t, err)

		// Verify expired refresh token is deleted
		var expiredToken auth.RefreshToken
		err = db.Where("token = ?", expiredTokenHash).First(&expiredToken).Error
		assert.Error(t, err) // Should not exist

		// Verify expired blacklist entry is deleted
		var expiredBlacklist auth.TokenBlacklist
		err = db.Where("jti = ?", expiredJTI).First(&expiredBlacklist).Error
		assert.Error(t, err) // Should not exist

		// Verify valid entries still exist
		var validToken auth.RefreshToken
		err = db.Where("token = ?", validTokenHash).First(&validToken).Error
		assert.NoError(t, err)

		var validBlacklist auth.TokenBlacklist
		err = db.Where("jti = ?", validJTI).First(&validBlacklist).Error
		assert.NoError(t, err)
	})
}
