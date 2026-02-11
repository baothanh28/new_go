package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"go.uber.org/zap"
	"myapp/internal/pkg/config"
)

// Service provides authentication business logic
type Service struct {
	userRepo        *Repository
	tokenRepo       *TokenRepository
	tokenManager    *TokenManager
	config          *config.Config
	logger          *zap.Logger
}

// NewService creates a new auth service
func NewService(
	userRepo *Repository,
	tokenRepo *TokenRepository,
	tokenManager *TokenManager,
	cfg *config.Config,
	logger *zap.Logger,
) *Service {
	return &Service{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		tokenManager: tokenManager,
		config:       cfg,
		logger:       logger,
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
	// Check if email already exists
	exists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("check email exists: %w", err)
	}
	if exists {
		return nil, &ErrEmailExists{Email: req.Email}
	}
	
	// Hash password
	hashedPassword, err := HashPassword(req.Password, &s.config.Auth)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	
	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = "user"
	}
	
	// Create user
	user := &User{
		Email:    req.Email,
		Password: hashedPassword,
		Role:     role,
	}
	
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	
	s.logger.Info("User registered successfully",
		zap.String("email", user.Email),
		zap.Uint("user_id", user.ID))
	
	return user, nil
}

// Login authenticates a user and returns access + refresh tokens
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warn("Login attempt with invalid email",
			zap.String("email", req.Email),
			zap.Error(err))
		return nil, &ErrInvalidCredentials{}
	}
	
	// Verify password
	if err := VerifyPassword(user.Password, req.Password); err != nil {
		s.logger.Warn("Login attempt with invalid password",
			zap.String("email", req.Email),
			zap.Uint("user_id", user.ID))
		return nil, &ErrInvalidCredentials{}
	}
	
	// Generate Access Token (RS256, 15 min)
	accessToken, err := s.tokenManager.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	
	// Generate Refresh Token (random string, 7 days)
	refreshToken, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	
	// Hash refresh token before storing
	refreshTokenHash := hashToken(refreshToken)
	expiresAt := time.Now().Add(s.tokenManager.GetRefreshTokenExpiration())
	
	// Store refresh token in database
	if err := s.tokenRepo.SaveRefreshToken(ctx, user.ID, refreshTokenHash, expiresAt); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}
	
	s.logger.Info("User logged in successfully",
		zap.String("email", user.Email),
		zap.Uint("user_id", user.ID))
	
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenManager.GetTokenExpiration().Seconds()),
		User:         user.ToUserResponse(),
	}, nil
}

// RefreshToken refreshes access token using refresh token (with rotation)
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*RefreshResponse, error) {
	// Hash the provided refresh token
	refreshTokenHash := hashToken(refreshToken)
	
	// Get refresh token from database
	storedToken, err := s.tokenRepo.GetRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	
	// Check if token is revoked
	if storedToken.Revoked {
		return nil, &ErrTokenRevoked{}
	}
	
	// Get user
	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	
	// Revoke old refresh token (token rotation)
	if err := s.tokenRepo.RevokeRefreshToken(ctx, refreshTokenHash); err != nil {
		s.logger.Warn("Failed to revoke old refresh token",
			zap.Uint("user_id", user.ID),
			zap.Error(err))
	}
	
	// Generate new Access Token
	newAccessToken, err := s.tokenManager.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	
	// Generate new Refresh Token
	newRefreshToken, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	
	// Hash and store new refresh token
	newRefreshTokenHash := hashToken(newRefreshToken)
	expiresAt := time.Now().Add(s.tokenManager.GetRefreshTokenExpiration())
	
	if err := s.tokenRepo.SaveRefreshToken(ctx, user.ID, newRefreshTokenHash, expiresAt); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}
	
	s.logger.Info("Token refreshed successfully",
		zap.Uint("user_id", user.ID))
	
	return &RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenManager.GetTokenExpiration().Seconds()),
	}, nil
}

// Logout revokes access token and all refresh tokens for a user
func (s *Service) Logout(ctx context.Context, accessToken string) error {
	// Validate and parse access token
	token, err := s.tokenManager.ValidateAccessToken(accessToken)
	if err != nil {
		return &ErrTokenInvalid{Message: err.Error()}
	}
	
	// Extract claims
	claims, err := s.tokenManager.ExtractClaims(token)
	if err != nil {
		return &ErrTokenInvalid{Message: err.Error()}
	}
	
	// Get JTI for blacklisting
	jti := claims.ID
	if jti == "" {
		return &ErrTokenInvalid{Message: "token missing JTI claim"}
	}
	
	// Add access token to blacklist
	expiresAt := claims.ExpiresAt.Time
	if err := s.tokenRepo.AddToBlacklist(ctx, jti, expiresAt); err != nil {
		s.logger.Warn("Failed to add token to blacklist",
			zap.String("jti", jti),
			zap.Error(err))
	}
	
	// Revoke all user's refresh tokens
	if err := s.tokenRepo.RevokeAllUserTokens(ctx, claims.UserID); err != nil {
		s.logger.Warn("Failed to revoke user refresh tokens",
			zap.Uint("user_id", claims.UserID),
			zap.Error(err))
	}
	
	s.logger.Info("User logged out successfully",
		zap.Uint("user_id", claims.UserID),
		zap.String("jti", jti))
	
	return nil
}

// ValidateToken validates an access token and returns claims
func (s *Service) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	// Validate token signature and expiration
	token, err := s.tokenManager.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, &ErrTokenInvalid{Message: err.Error()}
	}
	
	// Extract claims
	claims, err := s.tokenManager.ExtractClaims(token)
	if err != nil {
		return nil, &ErrTokenInvalid{Message: err.Error()}
	}
	
	// Check if token is blacklisted
	isBlacklisted, err := s.tokenRepo.IsBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("check blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, &ErrTokenRevoked{}
	}
	
	return claims, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, id uint) (*User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// hashToken hashes a token using SHA-256 for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
