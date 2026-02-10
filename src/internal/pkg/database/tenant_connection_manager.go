package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TenantConnectionManager manages dynamic database connections for tenants
type TenantConnectionManager struct {
	masterDB *gorm.DB
	logger   *zap.Logger
}

// NewTenantConnectionManager creates a new tenant connection manager
func NewTenantConnectionManager(masterDB *gorm.DB, logger *zap.Logger) *TenantConnectionManager {
	return &TenantConnectionManager{
		masterDB: masterDB,
		logger:   logger,
	}
}

// GetTenantDB retrieves or creates a database connection for the specified tenant
func (m *TenantConnectionManager) GetTenantDB(ctx context.Context, tenantID string) (*gorm.DB, error) {
	// Query master database for tenant configuration
	var tenant Tenant
	if err := m.masterDB.WithContext(ctx).Where("id = ? AND is_active = ?", tenantID, true).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant %s not found or inactive", tenantID)
		}
		return nil, fmt.Errorf("query tenant %s: %w", tenantID, err)
	}

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Silent)

	// Create database connection based on tenant's database type
	var dialector gorm.Dialector
	var dsn string

	switch tenant.DBType {
	case "postgresql", "postgres":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			tenant.DBHost, tenant.DBPort, tenant.DBUser, tenant.DBPassword, tenant.DBName)
		dialector = postgres.Open(dsn)

	case "mysql":
		// MySQL DSN format: user:password@tcp(host:port)/dbname?parseTime=true
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC",
			tenant.DBUser, tenant.DBPassword, tenant.DBHost, tenant.DBPort, tenant.DBName)
		dialector = mysql.Open(dsn)

	case "sqlite":
		// SQLite uses file path as DSN
		dsn = tenant.DBName // DBName should contain the file path for SQLite
		dialector = sqlite.Open(dsn)

	default:
		return nil, fmt.Errorf("unsupported database type '%s' for tenant %s", tenant.DBType, tenantID)
	}

	// Open database connection
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open %s database %s: %w", tenant.DBType, tenant.DBName, err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying database connection for tenant %s: %w", tenantID, err)
	}

	// Configure connection pool with reasonable defaults
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping %s database %s: %w", tenant.DBType, tenant.DBName, err)
	}

	m.logger.Debug("Tenant database connection established",
		zap.String("tenant_id", tenantID),
		zap.String("db_type", tenant.DBType),
		zap.String("db_name", tenant.DBName))

	return db, nil
}

// GetTenantConfig retrieves tenant configuration from master database
func (m *TenantConnectionManager) GetTenantConfig(ctx context.Context, tenantID string) (*Tenant, error) {
	var tenant Tenant
	if err := m.masterDB.WithContext(ctx).Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant %s not found", tenantID)
		}
		return nil, fmt.Errorf("query tenant %s: %w", tenantID, err)
	}
	return &tenant, nil
}
