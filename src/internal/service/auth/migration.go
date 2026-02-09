package auth

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunMigrations runs database migrations for auth service
func RunMigrations(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Running auth service migrations")
	
	if err := db.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("migrate users table: %w", err)
	}
	
	logger.Info("Auth service migrations completed successfully")
	return nil
}
