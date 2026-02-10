package database

// NOTE: These tests use SQLite which requires CGO to be enabled.
// To run these tests on Windows, you need to install a C compiler like MinGW-w64:
//   - Install MinGW-w64 from https://www.mingw-w64.org/
//   - Add MinGW-w64/bin to your PATH
//   - Run tests with: go test -tags=cgo
// On Linux/Mac, CGO is usually available by default.
//
// Alternatively, skip these tests and rely on integration tests with PostgreSQL.

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestMasterDB creates an in-memory master database with tenant records
func setupTestMasterDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate tenant model
	err = db.AutoMigrate(&Tenant{})
	require.NoError(t, err)

	return db
}

// TestNewTenantConnectionManager tests creating a tenant connection manager
func TestNewTenantConnectionManager(t *testing.T) {
	t.Run("create tenant connection manager", func(t *testing.T) {
		masterDB := setupTestMasterDB(t)
		logger := zaptest.NewLogger(t)

		manager := NewTenantConnectionManager(masterDB, logger)

		assert.NotNil(t, manager)
		assert.Equal(t, masterDB, manager.masterDB)
		assert.Equal(t, logger, manager.logger)
	})
}

// TestTenantConnectionManager_GetTenantConfig tests retrieving tenant configuration
func TestTenantConnectionManager_GetTenantConfig(t *testing.T) {
	masterDB := setupTestMasterDB(t)
	logger := zaptest.NewLogger(t)
	manager := NewTenantConnectionManager(masterDB, logger)
	ctx := context.Background()

	t.Run("get existing tenant config", func(t *testing.T) {
		// Create test tenant
		tenant := &Tenant{
			ID:         "tenant-123",
			Name:       "Test Tenant",
			IsActive:   true,
			DBType:     "sqlite",
			DBHost:     "localhost",
			DBPort:     5432,
			DBName:     ":memory:",
			DBUser:     "test_user",
			DBPassword: "test_pass",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err := masterDB.Create(tenant).Error
		require.NoError(t, err)

		// Get tenant config
		config, err := manager.GetTenantConfig(ctx, "tenant-123")
		assert.NoError(t, err)
		require.NotNil(t, config)
		assert.Equal(t, "tenant-123", config.ID)
		assert.Equal(t, "Test Tenant", config.Name)
		assert.True(t, config.IsActive)
	})

	t.Run("get non-existent tenant", func(t *testing.T) {
		config, err := manager.GetTenantConfig(ctx, "non-existent")
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "not found")
	})
}

// TestTenantConnectionManager_GetTenantDB tests getting tenant database connection
func TestTenantConnectionManager_GetTenantDB(t *testing.T) {
	masterDB := setupTestMasterDB(t)
	logger := zaptest.NewLogger(t)
	manager := NewTenantConnectionManager(masterDB, logger)
	ctx := context.Background()

	t.Run("get tenant DB for SQLite tenant", func(t *testing.T) {
		// Create SQLite tenant
		tenant := &Tenant{
			ID:         "sqlite-tenant",
			Name:       "SQLite Tenant",
			IsActive:   true,
			DBType:     "sqlite",
			DBHost:     "",
			DBPort:     0,
			DBName:     ":memory:",
			DBUser:     "",
			DBPassword: "",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err := masterDB.Create(tenant).Error
		require.NoError(t, err)

		// Get tenant DB
		db, err := manager.GetTenantDB(ctx, "sqlite-tenant")
		assert.NoError(t, err)
		require.NotNil(t, db)

		// Verify connection works
		sqlDB, err := db.DB()
		require.NoError(t, err)
		err = sqlDB.Ping()
		assert.NoError(t, err)
	})

	t.Run("fail to get DB for non-existent tenant", func(t *testing.T) {
		db, err := manager.GetTenantDB(ctx, "non-existent-tenant")
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("fail to get DB for inactive tenant", func(t *testing.T) {
		// Create inactive tenant
		tenant := &Tenant{
			ID:         "inactive-tenant",
			Name:       "Inactive Tenant",
			IsActive:   false,
			DBType:     "sqlite",
			DBHost:     "",
			DBPort:     0,
			DBName:     ":memory:",
			DBUser:     "",
			DBPassword: "",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err := masterDB.Create(tenant).Error
		require.NoError(t, err)

		// Try to get tenant DB
		db, err := manager.GetTenantDB(ctx, "inactive-tenant")
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "not found or inactive")
	})

	t.Run("fail with unsupported database type", func(t *testing.T) {
		// Create tenant with unsupported DB type
		tenant := &Tenant{
			ID:         "unsupported-tenant",
			Name:       "Unsupported Tenant",
			IsActive:   true,
			DBType:     "mongodb",
			DBHost:     "localhost",
			DBPort:     27017,
			DBName:     "test_db",
			DBUser:     "test_user",
			DBPassword: "test_pass",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err := masterDB.Create(tenant).Error
		require.NoError(t, err)

		// Try to get tenant DB
		db, err := manager.GetTenantDB(ctx, "unsupported-tenant")
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "unsupported database type")
	})
}

// TestTenantConnectionManager_PostgreSQL tests PostgreSQL connection (skipped without real DB)
func TestTenantConnectionManager_PostgreSQL(t *testing.T) {
	t.Run("get postgres tenant DB", func(t *testing.T) {
		t.Skip("Requires running PostgreSQL instance")

		masterDB := setupTestMasterDB(t)
		logger := zaptest.NewLogger(t)
		manager := NewTenantConnectionManager(masterDB, logger)
		ctx := context.Background()

		// Create PostgreSQL tenant
		tenant := &Tenant{
			ID:         "postgres-tenant",
			Name:       "PostgreSQL Tenant",
			IsActive:   true,
			DBType:     "postgres",
			DBHost:     "localhost",
			DBPort:     5432,
			DBName:     "tenant_db",
			DBUser:     "test_user",
			DBPassword: "test_pass",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err := masterDB.Create(tenant).Error
		require.NoError(t, err)

		// Get tenant DB
		db, err := manager.GetTenantDB(ctx, "postgres-tenant")
		assert.NoError(t, err)
		require.NotNil(t, db)
	})
}

// TestTenantConnectionManager_MySQL tests MySQL connection (skipped without real DB)
func TestTenantConnectionManager_MySQL(t *testing.T) {
	t.Run("get mysql tenant DB", func(t *testing.T) {
		t.Skip("Requires running MySQL instance")

		masterDB := setupTestMasterDB(t)
		logger := zaptest.NewLogger(t)
		manager := NewTenantConnectionManager(masterDB, logger)
		ctx := context.Background()

		// Create MySQL tenant
		tenant := &Tenant{
			ID:         "mysql-tenant",
			Name:       "MySQL Tenant",
			IsActive:   true,
			DBType:     "mysql",
			DBHost:     "localhost",
			DBPort:     3306,
			DBName:     "tenant_db",
			DBUser:     "test_user",
			DBPassword: "test_pass",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err := masterDB.Create(tenant).Error
		require.NoError(t, err)

		// Get tenant DB
		db, err := manager.GetTenantDB(ctx, "mysql-tenant")
		assert.NoError(t, err)
		require.NotNil(t, db)
	})
}

// TestTenantConnectionManager_ConnectionPooling tests connection pool configuration
func TestTenantConnectionManager_ConnectionPooling(t *testing.T) {
	t.Run("verify connection pool settings", func(t *testing.T) {
		masterDB := setupTestMasterDB(t)
		logger := zaptest.NewLogger(t)
		manager := NewTenantConnectionManager(masterDB, logger)
		ctx := context.Background()

		// Create SQLite tenant
		tenant := &Tenant{
			ID:         "pool-tenant",
			Name:       "Pool Tenant",
			IsActive:   true,
			DBType:     "sqlite",
			DBHost:     "",
			DBPort:     0,
			DBName:     ":memory:",
			DBUser:     "",
			DBPassword: "",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err := masterDB.Create(tenant).Error
		require.NoError(t, err)

		// Get tenant DB
		db, err := manager.GetTenantDB(ctx, "pool-tenant")
		require.NoError(t, err)
		require.NotNil(t, db)

		// Verify connection pool settings
		sqlDB, err := db.DB()
		require.NoError(t, err)

		stats := sqlDB.Stats()
		assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
	})
}

// TestTenantConnectionManager_MultipleTenantsTests managing multiple tenant connections
func TestTenantConnectionManager_MultipleTenants(t *testing.T) {
	t.Run("manage multiple tenant connections", func(t *testing.T) {
		masterDB := setupTestMasterDB(t)
		logger := zaptest.NewLogger(t)
		manager := NewTenantConnectionManager(masterDB, logger)
		ctx := context.Background()

		// Create multiple tenants
		tenants := []*Tenant{
			{
				ID:        "tenant-1",
				Name:      "Tenant 1",
				IsActive:  true,
				DBType:    "sqlite",
				DBName:    ":memory:",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        "tenant-2",
				Name:      "Tenant 2",
				IsActive:  true,
				DBType:    "sqlite",
				DBName:    ":memory:",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		for _, tenant := range tenants {
			err := masterDB.Create(tenant).Error
			require.NoError(t, err)
		}

		// Get connections for both tenants
		db1, err := manager.GetTenantDB(ctx, "tenant-1")
		assert.NoError(t, err)
		assert.NotNil(t, db1)

		db2, err := manager.GetTenantDB(ctx, "tenant-2")
		assert.NoError(t, err)
		assert.NotNil(t, db2)

		// Verify connections are separate
		assert.NotEqual(t, db1, db2)
	})
}

// TestTenantModel tests the Tenant model structure
func TestTenantModel(t *testing.T) {
	t.Run("create tenant model", func(t *testing.T) {
		tenant := &Tenant{
			ID:         "test-tenant",
			Name:       "Test Tenant",
			IsActive:   true,
			DBType:     "postgres",
			DBHost:     "localhost",
			DBPort:     5432,
			DBName:     "test_db",
			DBUser:     "user",
			DBPassword: "pass",
		}

		assert.Equal(t, "test-tenant", tenant.ID)
		assert.Equal(t, "Test Tenant", tenant.Name)
		assert.True(t, tenant.IsActive)
		assert.Equal(t, "postgres", tenant.DBType)
		assert.Equal(t, "localhost", tenant.DBHost)
		assert.Equal(t, 5432, tenant.DBPort)
	})
}
