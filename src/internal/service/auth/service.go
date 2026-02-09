package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"myapp/internal/pkg/config"
)

// Service provides authentication business logic
type Service struct {
	repo   *Repository
	cfg    *config.Config
	logger *zap.Logger
}

// NewService creates a new auth service
func NewService(repo *Repository, cfg *config.Config, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		cfg:    cfg,
		logger: logger,
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
	// Check if email already exists
	exists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("check email exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}
	
	// Hash password
	hashedPassword, err := hashPassword(req.Password)
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
	
	if err := s.repo.Insert(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	
	s.logger.Info("User registered successfully", zap.String("email", user.Email))
	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Get user by email
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	
	// Verify password
	if err := verifyPassword(user.Password, req.Password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	
	// Generate JWT token
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	
	s.logger.Info("User logged in successfully", zap.String("email", user.Email))
	
	return &LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// GenerateToken generates a JWT token for a user
func (s *Service) GenerateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * time.Duration(s.cfg.JWT.ExpirationHours)).Unix(),
		"iat":     time.Now().Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	
	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, fmt.Errorf("invalid token")
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, id uint) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id %d: %w", id, err)
	}
	return user, nil
}

// hashPassword hashes a plain text password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(bytes), nil
}

// verifyPassword verifies a plain text password against a hashed password
func verifyPassword(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return fmt.Errorf("password mismatch")
	}
	return nil
}
