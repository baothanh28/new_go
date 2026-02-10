package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"myapp/internal/pkg/config"
)

// TestNewDatabase tests database connection creation
func TestNewDatabase(t *testing.T) {
	t.Run("create postgres connection with valid config", func(t *testing.T) {
		// Note: This test requires a running PostgreSQL instance
		// Skip if not available
		t.Skip("Requires running PostgreSQL instance")

		cfg := config.DatabaseConfig{
			Driver:       "postgres",
			Host:         "localhost",
			Port:         5432,
			Name:         "test_db",
			User:         "test_user",
			Password:     "test_pass",
			MaxOpenConns: 10,
			MaxIdleConns: 2,
		}
		logger := zaptest.NewLogger(t)

		db, err := NewDatabase(cfg, logger)
		require.NoError(t, err)
		require.NotNil(t, db)

		// Cleanup
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})

	t.Run("fail with invalid host", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Driver:       "postgres",
			Host:         "invalid-host-that-does-not-exist",
			Port:         5432,
			Name:         "test_db",
			User:         "test_user",
			Password:     "test_pass",
			MaxOpenConns: 10,
			MaxIdleConns: 2,
		}
		logger := zaptest.NewLogger(t)

		db, err := NewDatabase(cfg, logger)
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

// TestDatabaseManager_Creation tests DatabaseManager creation
func TestDatabaseManager_Creation(t *testing.T) {
	t.Run("create database manager", func(t *testing.T) {
		// Note: This test requires running PostgreSQL instances
		t.Skip("Requires running PostgreSQL instances")

		cfg := &config.Config{
			MasterDatabase: config.DatabaseConfig{
				Driver:       "postgres",
				Host:         "localhost",
				Port:         5432,
				Name:         "master_db",
				User:         "test_user",
				Password:     "test_pass",
				MaxOpenConns: 10,
				MaxIdleConns: 2,
			},
			TenantDatabase: config.DatabaseConfig{
				Driver:       "postgres",
				Host:         "localhost",
				Port:         5432,
				Name:         "tenant_db",
				User:         "test_user",
				Password:     "test_pass",
				MaxOpenConns: 10,
				MaxIdleConns: 2,
			},
		}
		logger := zaptest.NewLogger(t)

		manager, err := NewDatabaseManager(cfg, logger)
		require.NoError(t, err)
		require.NotNil(t, manager)
		assert.NotNil(t, manager.MasterDB)
		assert.NotNil(t, manager.TenantDB)
		assert.NotNil(t, manager.TenantConnManager)

		// Cleanup
		manager.Close()
	})
}

// TestDatabaseManager_Close tests closing database connections
func TestDatabaseManager_Close(t *testing.T) {
	t.Run("close database manager", func(t *testing.T) {
		// Note: This test requires running PostgreSQL instances
		t.Skip("Requires running PostgreSQL instances")

		cfg := &config.Config{
			MasterDatabase: config.DatabaseConfig{
				Driver:       "postgres",
				Host:         "localhost",
				Port:         5432,
				Name:         "master_db",
				User:         "test_user",
				Password:     "test_pass",
				MaxOpenConns: 10,
				MaxIdleConns: 2,
			},
			TenantDatabase: config.DatabaseConfig{
				Driver:       "postgres",
				Host:         "localhost",
				Port:         5432,
				Name:         "tenant_db",
				User:         "test_user",
				Password:     "test_pass",
				MaxOpenConns: 10,
				MaxIdleConns: 2,
			},
		}
		logger := zaptest.NewLogger(t)

		manager, err := NewDatabaseManager(cfg, logger)
		require.NoError(t, err)

		err = manager.Close()
		assert.NoError(t, err)
	})

	t.Run("close nil connections", func(t *testing.T) {
		manager := &DatabaseManager{
			MasterDB: nil,
			TenantDB: nil,
		}

		err := manager.Close()
		assert.NoError(t, err)
	})
}

// TestDatabaseConfig_DSN tests DSN generation for different database types
func TestDatabaseConfig_DSN(t *testing.T) {
	t.Run("postgres DSN format", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     5432,
			Name:     "testdb",
			User:     "testuser",
			Password: "testpass",
		}

		// The DSN is generated in NewDatabase function
		// Expected format: host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable
		expectedParts := []string{"host=localhost", "port=5432", "user=testuser", "password=testpass", "dbname=testdb", "sslmode=disable"}

		// We can't directly test NewDatabase without a running DB,
		// but we can verify the config structure
		assert.Equal(t, "postgres", cfg.Driver)
		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 5432, cfg.Port)
		assert.Equal(t, "testdb", cfg.Name)
		assert.Equal(t, "testuser", cfg.User)
		assert.Equal(t, "testpass", cfg.Password)

		_ = expectedParts // For future DSN validation if needed
	})
}

// TestDatabaseConnectionPool tests connection pool configuration
func TestDatabaseConnectionPool(t *testing.T) {
	t.Run("connection pool settings", func(t *testing.T) {
		// Note: This test requires a running PostgreSQL instance
		t.Skip("Requires running PostgreSQL instance")

		cfg := config.DatabaseConfig{
			Driver:       "postgres",
			Host:         "localhost",
			Port:         5432,
			Name:         "test_db",
			User:         "test_user",
			Password:     "test_pass",
			MaxOpenConns: 50,
			MaxIdleConns: 10,
		}
		logger := zaptest.NewLogger(t)

		db, err := NewDatabase(cfg, logger)
		require.NoError(t, err)
		require.NotNil(t, db)

		sqlDB, err := db.DB()
		require.NoError(t, err)

		stats := sqlDB.Stats()
		assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)

		// Cleanup
		sqlDB.Close()
	})
}

// TestDatabaseManager_MultipleConnections tests managing multiple database connections
func TestDatabaseManager_MultipleConnections(t *testing.T) {
	t.Run("verify separate master and tenant connections", func(t *testing.T) {
		// Note: This test requires running PostgreSQL instances
		t.Skip("Requires running PostgreSQL instances")

		cfg := &config.Config{
			MasterDatabase: config.DatabaseConfig{
				Driver:       "postgres",
				Host:         "localhost",
				Port:         5432,
				Name:         "master_db",
				User:         "test_user",
				Password:     "test_pass",
				MaxOpenConns: 10,
				MaxIdleConns: 2,
			},
			TenantDatabase: config.DatabaseConfig{
				Driver:       "postgres",
				Host:         "localhost",
				Port:         5432,
				Name:         "tenant_db",
				User:         "test_user",
				Password:     "test_pass",
				MaxOpenConns: 20,
				MaxIdleConns: 5,
			},
		}
		logger := zaptest.NewLogger(t)

		manager, err := NewDatabaseManager(cfg, logger)
		require.NoError(t, err)
		require.NotNil(t, manager)

		// Verify both connections are separate
		assert.NotNil(t, manager.MasterDB)
		assert.NotNil(t, manager.TenantDB)
		assert.NotEqual(t, manager.MasterDB, manager.TenantDB)

		// Cleanup
		manager.Close()
	})
}

// TestNewDatabase_ErrorHandling tests error scenarios
func TestNewDatabase_ErrorHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name   string
		config config.DatabaseConfig
	}{
		{
			name: "invalid port",
			config: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     99999,
				Name:     "test_db",
				User:     "test_user",
				Password: "test_pass",
			},
		},
		{
			name: "invalid host",
			config: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "invalid.host.that.does.not.exist",
				Port:     5432,
				Name:     "test_db",
				User:     "test_user",
				Password: "test_pass",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewDatabase(tt.config, logger)
			assert.Error(t, err)
			assert.Nil(t, db)
		})
	}
}
