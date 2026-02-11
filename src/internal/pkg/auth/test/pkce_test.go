package auth_test

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"myapp/internal/pkg/auth"
)

func TestGenerateCodeChallenge_Plain(t *testing.T) {
	verifier := "test_verifier_123"
	challenge := auth.GenerateCodeChallenge(verifier, "plain")

	// Plain method should be base64url encoded verifier
	expected := base64.RawURLEncoding.EncodeToString([]byte(verifier))
	assert.Equal(t, expected, challenge)
}

func TestGenerateCodeChallenge_S256(t *testing.T) {
	verifier := "test_verifier_123"
	challenge := auth.GenerateCodeChallenge(verifier, "S256")

	// S256 method should be SHA-256 hash then base64url encoded
	hash := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(hash[:])
	assert.Equal(t, expected, challenge)
}

func TestGenerateCodeChallenge_Default(t *testing.T) {
	verifier := "test_verifier_123"
	challenge := auth.GenerateCodeChallenge(verifier, "invalid_method")

	// Default should use S256
	hash := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(hash[:])
	assert.Equal(t, expected, challenge)
}

func TestValidateCodeVerifier_Plain(t *testing.T) {
	verifier := "test_verifier_123"
	challenge := auth.GenerateCodeChallenge(verifier, "plain")

	tests := []struct {
		name      string
		verifier  string
		challenge string
		method    string
		want      bool
	}{
		{
			name:      "valid plain verifier",
			verifier:  verifier,
			challenge: challenge,
			method:    "plain",
			want:      true,
		},
		{
			name:      "invalid plain verifier",
			verifier:  "wrong_verifier",
			challenge: challenge,
			method:    "plain",
			want:      false,
		},
		{
			name:      "empty verifier",
			verifier:  "",
			challenge: auth.GenerateCodeChallenge("", "plain"),
			method:    "plain",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.ValidateCodeVerifier(tt.verifier, tt.challenge, tt.method)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateCodeVerifier_S256(t *testing.T) {
	verifier := "test_verifier_123"
	challenge := auth.GenerateCodeChallenge(verifier, "S256")

	tests := []struct {
		name      string
		verifier  string
		challenge string
		method    string
		want      bool
	}{
		{
			name:      "valid S256 verifier",
			verifier:  verifier,
			challenge: challenge,
			method:    "S256",
			want:      true,
		},
		{
			name:      "invalid S256 verifier",
			verifier:  "wrong_verifier",
			challenge: challenge,
			method:    "S256",
			want:      false,
		},
		{
			name:      "valid S256 with different verifier",
			verifier:  "another_verifier",
			challenge: auth.GenerateCodeChallenge("another_verifier", "S256"),
			method:    "S256",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.ValidateCodeVerifier(tt.verifier, tt.challenge, tt.method)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGenerateCodeVerifier(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "valid length 43",
			length:  43,
			wantErr: false,
		},
		{
			name:    "valid length 128",
			length:  128,
			wantErr: false,
		},
		{
			name:    "invalid length too short",
			length:  42,
			wantErr: true,
		},
		{
			name:    "invalid length too long",
			length:  129,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier, err := auth.GenerateCodeVerifier(tt.length)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, verifier)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, verifier)
				// Verify it's valid base64url
				decoded, err := base64.RawURLEncoding.DecodeString(verifier)
				assert.NoError(t, err)
				assert.NotEmpty(t, decoded)
			}
		})
	}
}

func TestGenerateCodeVerifier_Uniqueness(t *testing.T) {
	verifier1, err1 := auth.GenerateCodeVerifier(64)
	require.NoError(t, err1)

	verifier2, err2 := auth.GenerateCodeVerifier(64)
	require.NoError(t, err2)

	// Generated verifiers should be different (high probability)
	assert.NotEqual(t, verifier1, verifier2)
}

func TestPKCE_EndToEnd(t *testing.T) {
	// Generate a code verifier
	verifier, err := auth.GenerateCodeVerifier(64)
	require.NoError(t, err)

	// Generate challenge using S256
	challenge := auth.GenerateCodeChallenge(verifier, "S256")

	// Validate the verifier against challenge
	valid := auth.ValidateCodeVerifier(verifier, challenge, "S256")
	assert.True(t, valid)

	// Wrong verifier should fail
	wrongVerifier := "wrong_verifier"
	valid = auth.ValidateCodeVerifier(wrongVerifier, challenge, "S256")
	assert.False(t, valid)
}
