package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"myapp/internal/pkg/config"
)

// HashPassword hashes a plain text password using bcrypt
// Uses the cost from config, defaulting to bcrypt.DefaultCost if not set
func HashPassword(password string, cfg *config.AuthConfig) (string, error) {
	cost := cfg.BCryptCost
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(bytes), nil
}

// VerifyPassword verifies a plain text password against a hashed password
func VerifyPassword(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return fmt.Errorf("password mismatch: %w", err)
	}
	return nil
}
