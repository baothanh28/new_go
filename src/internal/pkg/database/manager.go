package database

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"myapp/internal/pkg/config"
)

// DatabaseManager manages master and tenant database connections
type DatabaseManager struct {
	MasterDB         *gorm.DB
	TenantDB         *gorm.DB // Deprecated: Use TenantConnManager for dynamic connections
	TenantConnManager *TenantConnectionManager
}

// NewDatabaseManager creates a new DatabaseManager with master and tenant connections
func NewDatabaseManager(cfg *config.Config, log *zap.Logger) (*DatabaseManager, error) {
	// Create master database connection
	masterDB, err := NewDatabase(cfg.MasterDatabase, log)
	if err != nil {
		return nil, fmt.Errorf("create master database connection: %w", err)
	}
	log.Info("Master database connected", 
		zap.String("host", cfg.MasterDatabase.Host),
		zap.String("name", cfg.MasterDatabase.Name))
	
	// Create tenant database connection (for backward compatibility)
	tenantDB, err := NewDatabase(cfg.TenantDatabase, log)
	if err != nil {
		return nil, fmt.Errorf("create tenant database connection: %w", err)
	}
	log.Info("Tenant database connected",
		zap.String("host", cfg.TenantDatabase.Host),
		zap.String("name", cfg.TenantDatabase.Name))
	
	// Create tenant connection manager for dynamic connections
	tenantConnManager := NewTenantConnectionManager(masterDB, log)
	log.Info("Tenant connection manager initialized")
	
	return &DatabaseManager{
		MasterDB:          masterDB,
		TenantDB:          tenantDB, // Kept for backward compatibility
		TenantConnManager: tenantConnManager,
	}, nil
}

// NewDatabase creates a new database connection based on configuration
func NewDatabase(cfg config.DatabaseConfig, log *zap.Logger) (*gorm.DB, error) {
	// For PostgreSQL only (as per user selection)
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
	
	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Silent)
	
	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres database %s: %w", cfg.Name, err)
	}
	
	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying database connection: %w", err)
	}
	
	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	
	log.Debug("Database connection pool configured",
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns))
	
	return db, nil
}

// Close closes all database connections
func (m *DatabaseManager) Close() error {
	var errors []error
	
	if m.MasterDB != nil {
		if sqlDB, err := m.MasterDB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errors = append(errors, fmt.Errorf("close master db: %w", err))
			}
		}
	}
	
	if m.TenantDB != nil {
		if sqlDB, err := m.TenantDB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errors = append(errors, fmt.Errorf("close tenant db: %w", err))
			}
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("errors closing databases: %v", errors)
	}
	
	return nil
}
