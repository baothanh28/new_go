package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"myapp/internal/pkg/auth"
	"myapp/internal/pkg/config"
)

func TestHashPassword(t *testing.T) {
	cfg := &config.AuthConfig{
		BCryptCost: 10, // Use lower cost for faster tests
	}

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "SecurePass123",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "VeryLongPassword12345678901234567890",
			wantErr:  false,
		},
		{
			name:     "short password",
			password: "short",
			wantErr:  false, // Hashing should succeed even for short passwords
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // Hashing should succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed, err := auth.HashPassword(tt.password, cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hashed)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hashed)
				assert.NotEqual(t, tt.password, hashed) // Should be hashed
				assert.Len(t, hashed, 60)                 // bcrypt hash is always 60 chars
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	cfg := &config.AuthConfig{
		BCryptCost: 10,
	}

	password := "SecurePass123"
	hashed, err := auth.HashPassword(password, cfg)
	require.NoError(t, err)

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		wantErr        bool
	}{
		{
			name:           "correct password",
			hashedPassword: hashed,
			password:       password,
			wantErr:        false,
		},
		{
			name:           "incorrect password",
			hashedPassword: hashed,
			password:       "WrongPassword",
			wantErr:        true,
		},
		{
			name:           "empty password",
			hashedPassword: hashed,
			password:       "",
			wantErr:        true,
		},
		{
			name:           "invalid hash",
			hashedPassword: "invalid_hash",
			password:       password,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.VerifyPassword(tt.hashedPassword, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHashPassword_DifferentPasswords(t *testing.T) {
	cfg := &config.AuthConfig{
		BCryptCost: 10,
	}

	password1 := "Password1"
	password2 := "Password2"

	hashed1, err1 := auth.HashPassword(password1, cfg)
	require.NoError(t, err1)

	hashed2, err2 := auth.HashPassword(password2, cfg)
	require.NoError(t, err2)

	// Different passwords should produce different hashes
	assert.NotEqual(t, hashed1, hashed2)

	// Each password should verify correctly
	assert.NoError(t, auth.VerifyPassword(hashed1, password1))
	assert.NoError(t, auth.VerifyPassword(hashed2, password2))

	// Cross-verification should fail
	assert.Error(t, auth.VerifyPassword(hashed1, password2))
	assert.Error(t, auth.VerifyPassword(hashed2, password1))
}

func TestHashPassword_SamePasswordDifferentHashes(t *testing.T) {
	cfg := &config.AuthConfig{
		BCryptCost: 10,
	}

	password := "SamePassword123"

	// Hash the same password multiple times
	hashed1, err1 := auth.HashPassword(password, cfg)
	require.NoError(t, err1)

	hashed2, err2 := auth.HashPassword(password, cfg)
	require.NoError(t, err2)

	// Same password should produce different hashes (due to salt)
	assert.NotEqual(t, hashed1, hashed2)

	// But both should verify correctly
	assert.NoError(t, auth.VerifyPassword(hashed1, password))
	assert.NoError(t, auth.VerifyPassword(hashed2, password))
}
