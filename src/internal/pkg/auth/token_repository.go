package auth

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"myapp/internal/pkg/database"
)

// TokenRepository provides database operations for refresh tokens and token blacklist
type TokenRepository struct {
	refreshTokenRepo *database.MasterRepo[RefreshToken]
	blacklistRepo    *database.MasterRepo[TokenBlacklist]
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(dbManager *database.DatabaseManager) *TokenRepository {
	return &TokenRepository{
		refreshTokenRepo: database.NewMasterRepo[RefreshToken](dbManager),
		blacklistRepo:    database.NewMasterRepo[TokenBlacklist](dbManager),
	}
}

// SaveRefreshToken saves a refresh token to the database (hashed)
func (r *TokenRepository) SaveRefreshToken(ctx context.Context, userID uint, tokenHash string, expiresAt time.Time) error {
	refreshToken := &RefreshToken{
		UserID:    userID,
		Token:     tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		Revoked:   false,
	}
	
	if err := r.refreshTokenRepo.Insert(ctx, refreshToken); err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

// GetRefreshToken retrieves a refresh token by its hash
func (r *TokenRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	var refreshToken RefreshToken
	if err := r.refreshTokenRepo.GetDB().WithContext(ctx).
		Where("token = ? AND revoked = ?", tokenHash, false).
		First(&refreshToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ErrRefreshTokenNotFound{}
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	
	// Check if token has expired
	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, &ErrTokenExpired{Message: "refresh token has expired"}
	}
	
	return &refreshToken, nil
}

// RevokeRefreshToken revokes a refresh token
func (r *TokenRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	now := time.Now()
	if err := r.refreshTokenRepo.GetDB().WithContext(ctx).
		Model(&RefreshToken{}).
		Where("token = ?", tokenHash).
		Updates(map[string]interface{}{
			"revoked":   true,
			"revoked_at": now,
		}).Error; err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (r *TokenRepository) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	now := time.Now()
	if err := r.refreshTokenRepo.GetDB().WithContext(ctx).
		Model(&RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Updates(map[string]interface{}{
			"revoked":   true,
			"revoked_at": now,
		}).Error; err != nil {
		return fmt.Errorf("revoke all user tokens: %w", err)
	}
	return nil
}

// AddToBlacklist adds a JTI to the token blacklist
func (r *TokenRepository) AddToBlacklist(ctx context.Context, jti string, expiresAt time.Time) error {
	blacklistEntry := &TokenBlacklist{
		JTI:       jti,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	
	if err := r.blacklistRepo.Insert(ctx, blacklistEntry); err != nil {
		// Ignore duplicate key errors (token already blacklisted)
		return nil
	}
	return nil
}

// IsBlacklisted checks if a JTI is in the blacklist
func (r *TokenRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	var count int64
	if err := r.blacklistRepo.GetDB().WithContext(ctx).
		Model(&TokenBlacklist{}).
		Where("jti = ?", jti).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check blacklist: %w", err)
	}
	return count > 0, nil
}

// CleanupExpiredTokens removes expired tokens from blacklist and refresh tokens
func (r *TokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	now := time.Now()
	
	// Cleanup expired blacklist entries
	if err := r.blacklistRepo.GetDB().WithContext(ctx).
		Where("expires_at < ?", now).
		Delete(&TokenBlacklist{}).Error; err != nil {
		return fmt.Errorf("cleanup expired blacklist: %w", err)
	}
	
	// Cleanup expired refresh tokens
	if err := r.refreshTokenRepo.GetDB().WithContext(ctx).
		Where("expires_at < ?", now).
		Delete(&RefreshToken{}).Error; err != nil {
		return fmt.Errorf("cleanup expired refresh tokens: %w", err)
	}
	
	return nil
}
