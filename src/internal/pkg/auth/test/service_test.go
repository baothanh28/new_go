// +build cgo

package auth_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"myapp/internal/pkg/auth"
	"myapp/internal/pkg/auth/keys"
	"myapp/internal/pkg/config"
	"myapp/internal/pkg/database"
)

// setupTestService creates a complete test service with all dependencies
func setupTestService(t *testing.T) (*auth.Service, func()) {
	// Setup database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&auth.User{}, &auth.RefreshToken{}, &auth.TokenBlacklist{})
	require.NoError(t, err)

	dbManager := &database.DatabaseManager{
		MasterDB: db,
	}

	// Setup repositories
	userRepo := auth.NewRepository(dbManager)
	tokenRepo := auth.NewTokenRepository(dbManager)

	// Setup token manager
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	err = keys.GenerateAndSaveKeyPair(privateKeyPath, publicKeyPath, 2048)
	require.NoError(t, err)

	authConfig := &config.AuthConfig{
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		RSAPrivateKeyPath:    privateKeyPath,
		RSAPublicKeyPath:     publicKeyPath,
		Issuer:               "test-issuer",
		BCryptCost:           10,
	}

	tokenManager, err := auth.NewTokenManager(authConfig)
	require.NoError(t, err)

	appConfig := &config.Config{
		Auth: *authConfig,
	}

	logger := zap.NewNop()

	service := auth.NewService(userRepo, tokenRepo, tokenManager, appConfig, logger)

	cleanup := func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
	}

	return service, cleanup
}

func TestService_Register(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("register new user", func(t *testing.T) {
		req := &auth.RegisterRequest{
			Email:    "newuser@example.com",
			Password: "SecurePass123",
			Role:     "user",
		}

		user, err := service.Register(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Email, user.Email)
		assert.Equal(t, req.Role, user.Role)
		assert.NotEmpty(t, user.Password) // Should be hashed
		assert.NotEqual(t, req.Password, user.Password)
	})

	t.Run("register duplicate email", func(t *testing.T) {
		req := &auth.RegisterRequest{
			Email:    "duplicate@example.com",
			Password: "SecurePass123",
		}

		_, err := service.Register(ctx, req)
		assert.NoError(t, err)

		// Try to register again
		_, err = service.Register(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &auth.ErrEmailExists{}, err)
	})

	t.Run("register with default role", func(t *testing.T) {
		req := &auth.RegisterRequest{
			Email:    "norole@example.com",
			Password: "SecurePass123",
		}

		user, err := service.Register(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, "user", user.Role) // Default role
	})
}

func TestService_Login(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()
	ctx := context.Background()

	// Register a user first
	registerReq := &RegisterRequest{
		Email:    "login@example.com",
		Password: "SecurePass123",
		Role:     "user",
	}
	_, err := service.Register(ctx, registerReq)
	require.NoError(t, err)

	t.Run("login with correct credentials", func(t *testing.T) {
		req := &auth.LoginRequest{
			Email:    "login@example.com",
			Password: "SecurePass123",
		}

		response, err := service.Login(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, "Bearer", response.TokenType)
		assert.Equal(t, int64(900), response.ExpiresIn) // 15 minutes in seconds
		assert.NotNil(t, response.User)
		assert.Equal(t, "login@example.com", response.User.Email)
	})

	t.Run("login with incorrect email", func(t *testing.T) {
		req := &auth.LoginRequest{
			Email:    "wrong@example.com",
			Password: "SecurePass123",
		}

		_, err := service.Login(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &auth.ErrInvalidCredentials{}, err)
	})

	t.Run("login with incorrect password", func(t *testing.T) {
		req := &auth.LoginRequest{
			Email:    "login@example.com",
			Password: "WrongPassword",
		}

		_, err := service.Login(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &auth.ErrInvalidCredentials{}, err)
	})
}

func TestService_RefreshToken(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()
	ctx := context.Background()

	// Register and login to get refresh token
	registerReq := &RegisterRequest{
		Email:    "refresh@example.com",
		Password: "SecurePass123",
	}
	_, err := service.Register(ctx, registerReq)
	require.NoError(t, err)

	loginReq := &LoginRequest{
		Email:    "refresh@example.com",
		Password: "SecurePass123",
	}
	loginResponse, err := service.Login(ctx, loginReq)
	require.NoError(t, err)
	require.NotEmpty(t, loginResponse.RefreshToken)

	t.Run("refresh token successfully", func(t *testing.T) {
		response, err := service.RefreshToken(ctx, loginResponse.RefreshToken)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.NotEqual(t, loginResponse.RefreshToken, response.RefreshToken) // Should rotate
		assert.Equal(t, "Bearer", response.TokenType)
	})

	t.Run("refresh with invalid token", func(t *testing.T) {
		_, err := service.RefreshToken(ctx, "invalid_refresh_token")
		assert.Error(t, err)
		assert.IsType(t, &auth.ErrRefreshTokenNotFound{}, err)
	})

	t.Run("refresh token rotation", func(t *testing.T) {
		// Login again to get a fresh token
		loginResponse2, err := service.Login(ctx, loginReq)
		require.NoError(t, err)

		// Refresh the token
		refreshResponse1, err := service.RefreshToken(ctx, loginResponse2.RefreshToken)
		require.NoError(t, err)

		// Try to use the old refresh token (should fail due to rotation)
		_, err = service.RefreshToken(ctx, loginResponse2.RefreshToken)
		assert.Error(t, err)

		// Use the new refresh token (should work)
		refreshResponse2, err := service.RefreshToken(ctx, refreshResponse1.RefreshToken)
		assert.NoError(t, err)
		assert.NotEqual(t, refreshResponse1.RefreshToken, refreshResponse2.RefreshToken)
	})
}

func TestService_Logout(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()
	ctx := context.Background()

	// Register and login
	registerReq := &RegisterRequest{
		Email:    "logout@example.com",
		Password: "SecurePass123",
	}
	_, err := service.Register(ctx, registerReq)
	require.NoError(t, err)

	loginReq := &LoginRequest{
		Email:    "logout@example.com",
		Password: "SecurePass123",
	}
	loginResponse, err := service.Login(ctx, loginReq)
	require.NoError(t, err)

	t.Run("logout successfully", func(t *testing.T) {
		err := service.Logout(ctx, loginResponse.AccessToken)
		assert.NoError(t, err)

		// Verify token is blacklisted
		_, err = service.ValidateToken(ctx, loginResponse.AccessToken)
		assert.Error(t, err)
		assert.IsType(t, &auth.ErrTokenRevoked{}, err)
	})

	t.Run("logout with invalid token", func(t *testing.T) {
		err := service.Logout(ctx, "invalid_token")
		assert.Error(t, err)
		assert.IsType(t, &auth.ErrTokenInvalid{}, err)
	})
}

func TestService_ValidateToken(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()
	ctx := context.Background()

	// Register and login
	registerReq := &RegisterRequest{
		Email:    "validate@example.com",
		Password: "SecurePass123",
	}
	_, err := service.Register(ctx, registerReq)
	require.NoError(t, err)

	loginReq := &LoginRequest{
		Email:    "validate@example.com",
		Password: "SecurePass123",
	}
	loginResponse, err := service.Login(ctx, loginReq)
	require.NoError(t, err)

	t.Run("validate valid token", func(t *testing.T) {
		claims, err := service.ValidateToken(ctx, loginResponse.AccessToken)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "validate@example.com", claims.Email)
	})

	t.Run("validate invalid token", func(t *testing.T) {
		_, err := service.ValidateToken(ctx, "invalid_token")
		assert.Error(t, err)
		assert.IsType(t, &auth.ErrTokenInvalid{}, err)
	})

	t.Run("validate blacklisted token", func(t *testing.T) {
		// Logout to blacklist token
		err := service.Logout(ctx, loginResponse.AccessToken)
		require.NoError(t, err)

		// Try to validate blacklisted token
		_, err = service.ValidateToken(ctx, loginResponse.AccessToken)
		assert.Error(t, err)
		assert.IsType(t, &auth.ErrTokenRevoked{}, err)
	})
}

func TestService_GetUserByID(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()
	ctx := context.Background()

	// Register a user
	registerReq := &RegisterRequest{
		Email:    "getuser@example.com",
		Password: "SecurePass123",
	}
	user, err := service.Register(ctx, registerReq)
	require.NoError(t, err)

	t.Run("get existing user", func(t *testing.T) {
		found, err := service.GetUserByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("get non-existent user", func(t *testing.T) {
		_, err := service.GetUserByID(ctx, 999)
		assert.Error(t, err)
	})
}
