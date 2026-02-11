package auth_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"myapp/internal/pkg/auth"
	"myapp/internal/pkg/auth/keys"
	"myapp/internal/pkg/config"
)

// setupTestTokenManager creates a TokenManager with temporary keys for testing
func setupTestTokenManager(t *testing.T) (*auth.TokenManager, func()) {
	// Create temporary directory for keys
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Generate keys
	err := keys.GenerateAndSaveKeyPair(privateKeyPath, publicKeyPath, 2048)
	require.NoError(t, err)

	cfg := &config.AuthConfig{
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		RSAPrivateKeyPath:    privateKeyPath,
		RSAPublicKeyPath:     publicKeyPath,
		Issuer:               "test-issuer",
		BCryptCost:           10,
	}

	tm, err := auth.NewTokenManager(cfg)
	require.NoError(t, err)

	cleanup := func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
	}

	return tm, cleanup
}

func TestNewTokenManager(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Generate keys
	err := keys.GenerateAndSaveKeyPair(privateKeyPath, publicKeyPath, 2048)
	require.NoError(t, err)

	cfg := &config.AuthConfig{
		RSAPrivateKeyPath: privateKeyPath,
		RSAPublicKeyPath:  publicKeyPath,
		Issuer:            "test-issuer",
	}

	tm, err := auth.NewTokenManager(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, tm)
}

func TestNewTokenManager_InvalidKeyPath(t *testing.T) {
	cfg := &config.AuthConfig{
		RSAPrivateKeyPath: "/nonexistent/private.pem",
		RSAPublicKeyPath:  "/nonexistent/public.pem",
		Issuer:            "test-issuer",
	}

	tm, err := auth.NewTokenManager(cfg)
	assert.Error(t, err)
	assert.Nil(t, tm)
}

func TestTokenManager_GenerateAccessToken(t *testing.T) {
	tm, cleanup := setupTestTokenManager(t)
	defer cleanup()

	user := &auth.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "user",
	}

	token, err := tm.GenerateAccessToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Token should be a valid JWT - use ValidateAccessToken instead
	parsed, err := tm.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.True(t, parsed.Valid)
}

func TestTokenManager_GenerateAccessToken_Claims(t *testing.T) {
	tm, cleanup := setupTestTokenManager(t)
	defer cleanup()

	user := &auth.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "admin",
	}

	token, err := tm.GenerateAccessToken(user)
	require.NoError(t, err)

	// Validate and extract claims
	validatedToken, err := tm.ValidateAccessToken(token)
	require.NoError(t, err)

	claims, err := tm.ExtractClaims(validatedToken)
	require.NoError(t, err)

	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
	assert.Equal(t, "test-issuer", claims.Issuer) // From setupTestTokenManager
	assert.NotEmpty(t, claims.ID) // JTI should be present
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
}

func TestTokenManager_GenerateRefreshToken(t *testing.T) {
	tm, cleanup := setupTestTokenManager(t)
	defer cleanup()

	token1, err := tm.GenerateRefreshToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := tm.GenerateRefreshToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Refresh tokens should be different
	assert.NotEqual(t, token1, token2)
}

func TestTokenManager_ValidateAccessToken(t *testing.T) {
	tm, cleanup := setupTestTokenManager(t)
	defer cleanup()

	user := &auth.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "user",
	}

	token, err := tm.GenerateAccessToken(user)
	require.NoError(t, err)

	// Valid token should validate successfully
	validatedToken, err := tm.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, validatedToken)
	assert.True(t, validatedToken.Valid)
}

func TestTokenManager_ValidateAccessToken_Invalid(t *testing.T) {
	tm, cleanup := setupTestTokenManager(t)
	defer cleanup()

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "invalid format",
			token: "invalid.token.format",
		},
		{
			name:  "wrong signature",
			token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tm.ValidateAccessToken(tt.token)
			assert.Error(t, err)
		})
	}
}

func TestTokenManager_ValidateAccessToken_Expired(t *testing.T) {
	// Create token manager with very short expiration
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	err := keys.GenerateAndSaveKeyPair(privateKeyPath, publicKeyPath, 2048)
	require.NoError(t, err)

	cfg := &config.AuthConfig{
		AccessTokenDuration:  1 * time.Nanosecond, // Very short expiration
		RefreshTokenDuration: 7 * 24 * time.Hour,
		RSAPrivateKeyPath:    privateKeyPath,
		RSAPublicKeyPath:     publicKeyPath,
		Issuer:               "test-issuer",
		BCryptCost:           10,
	}

	tm, err := auth.NewTokenManager(cfg)
	require.NoError(t, err)

	user := &auth.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "user",
	}

	token, err := tm.GenerateAccessToken(user)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Expired token should fail validation
	_, err = tm.ValidateAccessToken(token)
	assert.Error(t, err)
}

func TestTokenManager_ExtractClaims(t *testing.T) {
	tm, cleanup := setupTestTokenManager(t)
	defer cleanup()

	user := &auth.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "user",
	}

	token, err := tm.GenerateAccessToken(user)
	require.NoError(t, err)

	validatedToken, err := tm.ValidateAccessToken(token)
	require.NoError(t, err)

	claims, err := tm.ExtractClaims(validatedToken)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
}

func TestTokenManager_GetTokenExpiration(t *testing.T) {
	tm, cleanup := setupTestTokenManager(t)
	defer cleanup()

	expiration := tm.GetTokenExpiration()
	assert.Equal(t, 15*time.Minute, expiration)
}

func TestTokenManager_GetRefreshTokenExpiration(t *testing.T) {
	tm, cleanup := setupTestTokenManager(t)
	defer cleanup()

	expiration := tm.GetRefreshTokenExpiration()
	assert.Equal(t, 7*24*time.Hour, expiration)
}

func TestTokenManager_DifferentUsers(t *testing.T) {
	tm, cleanup := setupTestTokenManager(t)
	defer cleanup()

	user1 := &auth.User{ID: 1, Email: "user1@example.com", Role: "user"}
	user2 := &auth.User{ID: 2, Email: "user2@example.com", Role: "admin"}

	token1, err1 := tm.GenerateAccessToken(user1)
	require.NoError(t, err1)

	token2, err2 := tm.GenerateAccessToken(user2)
	require.NoError(t, err2)

	// Tokens should be different
	assert.NotEqual(t, token1, token2)

	// Extract and verify claims
	validatedToken1, err := tm.ValidateAccessToken(token1)
	require.NoError(t, err)
	claims1, err := tm.ExtractClaims(validatedToken1)
	require.NoError(t, err)

	validatedToken2, err := tm.ValidateAccessToken(token2)
	require.NoError(t, err)
	claims2, err := tm.ExtractClaims(validatedToken2)
	require.NoError(t, err)

	assert.Equal(t, user1.ID, claims1.UserID)
	assert.Equal(t, user2.ID, claims2.UserID)
	assert.NotEqual(t, claims1.ID, claims2.ID) // Different JTIs
}
