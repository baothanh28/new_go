package auth

import (
	"fmt"
	"time"
)

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents user login response with tokens
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"` // "Bearer"
	ExpiresIn    int64        `json:"expires_in"` // seconds until access token expires
	User         UserResponse `json:"user"`
}

// RefreshRequest represents refresh token request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshResponse represents refresh token response
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// UserResponse represents user data response (without sensitive info)
type UserResponse struct {
	ID        string    `json:"id"` // UUIDv7
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ToUserResponse converts a User model to UserResponse DTO
// Note: User.ID should be updated to use UUIDv7 (string type) in the User model
func (u *User) ToUserResponse() UserResponse {
	// Convert ID to string - if User.ID is already UUID string, this will work
	// If User.ID is uint, it will be converted to string (consider updating User model to use UUIDv7)
	var idStr string
	if u.ID != 0 {
		idStr = fmt.Sprintf("%d", u.ID)
	}
	return UserResponse{
		ID:        idStr,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}
