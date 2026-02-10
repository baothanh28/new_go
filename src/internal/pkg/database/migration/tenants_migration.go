package migration

import (
	"fmt"
	"gorm.io/gorm"
	"myapp/internal/pkg/database"
)

// MigrateTenants creates the tenants table in the master database
func MigrateTenants(db *gorm.DB) error {
	if err := db.AutoMigrate(&database.Tenant{}); err != nil {
		return fmt.Errorf("migrate tenants table: %w", err)
	}
	return nil
}

// SeedSampleTenant seeds a sample tenant for testing (optional)
func SeedSampleTenant(db *gorm.DB) error {
	// Check if tenant already exists
	var count int64
	if err := db.Model(&database.Tenant{}).Where("id = ?", "tenant-001").Count(&count).Error; err != nil {
		return fmt.Errorf("check existing tenant: %w", err)
	}
	
	if count > 0 {
		return nil // Tenant already exists
	}
	
	// Create sample tenant
	tenant := &database.Tenant{
		ID:         "tenant-001",
		Name:       "Sample Tenant",
		DBHost:     "localhost",
		DBPort:     5432,
		DBName:     "tenant_001_db",
		DBUser:     "postgres",
		DBPassword: "password",
		IsActive:   true,
	}
	
	if err := db.Create(tenant).Error; err != nil {
		return fmt.Errorf("seed sample tenant: %w", err)
	}
	
	return nil
}
