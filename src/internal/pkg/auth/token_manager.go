package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"myapp/internal/pkg/auth/keys"
	"myapp/internal/pkg/config"
)

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenManager handles JWT token generation and validation using RS256
type TokenManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	config     *config.AuthConfig
}

// NewTokenManager creates a new TokenManager instance
func NewTokenManager(cfg *config.AuthConfig) (*TokenManager, error) {
	// Load RSA private key
	privateKey, err := keys.LoadPrivateKeyPEM(cfg.RSAPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}
	
	// Load RSA public key
	publicKey, err := keys.LoadPublicKeyPEM(cfg.RSAPublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load public key: %w", err)
	}
	
	return &TokenManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		config:     cfg,
	}, nil
}

// GenerateAccessToken generates a new JWT access token for a user (RS256)
// Token expires after AccessTokenDuration (default: 15 minutes)
func (tm *TokenManager) GenerateAccessToken(user *User) (string, error) {
	now := time.Now()
	expiresAt := now.Add(tm.config.AccessTokenDuration)
	
	// Generate unique JTI (JWT ID) for token revocation
	jti := generateJTI()
	
	claims := &TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    tm.config.Issuer,
			ID:        jti,
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(tm.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	
	return tokenString, nil
}

// GenerateRefreshToken generates a secure random refresh token string
// This is not a JWT, just a random string that will be hashed and stored
func (tm *TokenManager) GenerateRefreshToken() (string, error) {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	
	// Base64url encode for URL-safe token
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// ValidateAccessToken validates and parses an access token
func (tm *TokenManager) ValidateAccessToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.publicKey, nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	
	return token, nil
}

// ExtractClaims extracts TokenClaims from a validated JWT token
func (tm *TokenManager) ExtractClaims(token *jwt.Token) (*TokenClaims, error) {
	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims type")
	}
	return claims, nil
}

// GetTokenExpiration returns the expiration time for access tokens
func (tm *TokenManager) GetTokenExpiration() time.Duration {
	return tm.config.AccessTokenDuration
}

// GetRefreshTokenExpiration returns the expiration time for refresh tokens
func (tm *TokenManager) GetRefreshTokenExpiration() time.Duration {
	return tm.config.RefreshTokenDuration
}

// generateJTI generates a unique JWT ID (JTI) for token revocation
func generateJTI() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}
