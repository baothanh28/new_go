package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// ValidateCodeVerifier validates PKCE code_verifier against code_challenge
// method: "plain" or "S256" (SHA-256)
func ValidateCodeVerifier(verifier, challenge, method string) bool {
	expectedChallenge := GenerateCodeChallenge(verifier, method)
	return expectedChallenge == challenge
}

// GenerateCodeChallenge generates code_challenge from verifier using the specified method
// method: "plain" or "S256" (SHA-256)
func GenerateCodeChallenge(verifier, method string) string {
	switch method {
	case "plain":
		// For plain method, return base64url encoded verifier
		return base64.RawURLEncoding.EncodeToString([]byte(verifier))
	case "S256":
		// For S256 method, hash with SHA-256 then base64url encode
		hash := sha256.Sum256([]byte(verifier))
		return base64.RawURLEncoding.EncodeToString(hash[:])
	default:
		// Default to S256 for security
		hash := sha256.Sum256([]byte(verifier))
		return base64.RawURLEncoding.EncodeToString(hash[:])
	}
}

// GenerateCodeVerifier generates a random code_verifier for PKCE
// Returns base64url encoded random string (43-128 characters)
func GenerateCodeVerifier(length int) (string, error) {
	if length < 43 || length > 128 {
		return "", fmt.Errorf("code verifier length must be between 43 and 128 characters")
	}
	
	// Generate random bytes
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	
	// Base64url encode
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}
