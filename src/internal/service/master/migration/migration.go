package migration

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"myapp/internal/service/master/model"
)

// RunMigrations runs database migrations for master service
func RunMigrations(db *gorm.DB, logger *zap.Logger) error {
	if err := db.AutoMigrate(&model.Master{}); err != nil {
		return fmt.Errorf("failed to migrate master table: %w", err)
	}
	
	logger.Info("Master table migrated successfully")
	
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	
	return nil
}

// createIndexes creates additional indexes for master table
func createIndexes(db *gorm.DB) error {
	// Indexes are already created via GORM tags:
	// - uniqueIndex on code
	// - index on type
	// - index on deleted_at (soft delete)
	// - index on is_active (if needed for filtering)
	
	// Add composite indexes if needed for common queries
	// Example: composite index on (type, is_active)
	// if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_masters_type_active ON masters(type, is_active)").Error; err != nil {
	//     return fmt.Errorf("create composite index: %w", err)
	// }
	
	return nil
}
