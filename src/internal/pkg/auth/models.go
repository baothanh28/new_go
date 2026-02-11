package auth

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"` // Never expose password in JSON
	Role      string    `gorm:"not null;default:'user'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// RefreshToken represents a refresh token for token rotation
type RefreshToken struct {
	ID        uint       `gorm:"primarykey"`
	UserID    uint       `gorm:"index;not null"`
	Token     string     `gorm:"uniqueIndex;not null"` // Hashed token value
	ExpiresAt time.Time  `gorm:"index;not null"`
	CreatedAt time.Time  `gorm:"not null"`
	Revoked   bool       `gorm:"default:false"`
	RevokedAt *time.Time `gorm:"default:null"`
}

// TableName specifies the table name for RefreshToken model
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// TokenBlacklist represents a blacklisted access token (by JTI)
type TokenBlacklist struct {
	JTI       string    `gorm:"primarykey"` // JWT ID (JTI claim)
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time `gorm:"not null"`
}

// TableName specifies the table name for TokenBlacklist model
func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}
