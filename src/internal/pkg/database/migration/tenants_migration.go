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
	
	// Create sample tenants with different database types
	tenants := []*database.Tenant{
		{
			ID:       "tenant-001",
			Name:     "PostgreSQL Tenant",
			DBType:   "postgresql",
			Cnn:      "host=localhost port=5432 user=postgres password=password dbname=tenant_001_db sslmode=disable",
			IsActive: true,
		},
		{
			ID:       "tenant-002",
			Name:     "MySQL Tenant",
			DBType:   "mysql",
			Cnn:      "mysqluser:mysqlpass@tcp(localhost:3307)/tenant_db_1?parseTime=true&loc=UTC&allowPublicKeyRetrieval=true",
			IsActive: true,
		},
	}
	
	for _, tenant := range tenants {
		if err := db.Create(tenant).Error; err != nil {
			return fmt.Errorf("seed tenant %s: %w", tenant.ID, err)
		}
	}
	
	return nil
}
