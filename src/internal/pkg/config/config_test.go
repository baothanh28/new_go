package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestServerConfig_Validate tests ServerConfig validation
func TestServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ServerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid server config",
			config: ServerConfig{
				Host: "0.0.0.0",
				Port: 8080,
			},
			wantErr: false,
		},
		{
			name: "valid server config with localhost",
			config: ServerConfig{
				Host: "localhost",
				Port: 3000,
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: ServerConfig{
				Host: "",
				Port: 8080,
			},
			wantErr: true,
			errMsg:  "server host is required",
		},
		{
			name: "invalid port - zero",
			config: ServerConfig{
				Host: "0.0.0.0",
				Port: 0,
			},
			wantErr: true,
			errMsg:  "server port must be between 1 and 65535",
		},
		{
			name: "invalid port - negative",
			config: ServerConfig{
				Host: "0.0.0.0",
				Port: -1,
			},
			wantErr: true,
			errMsg:  "server port must be between 1 and 65535",
		},
		{
			name: "invalid port - too high",
			config: ServerConfig{
				Host: "0.0.0.0",
				Port: 65536,
			},
			wantErr: true,
			errMsg:  "server port must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDatabaseConfig_Validate tests DatabaseConfig validation
func TestDatabaseConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  DatabaseConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid postgres config",
			config: DatabaseConfig{
				Driver:       "postgres",
				Host:         "localhost",
				Port:         5432,
				Name:         "mydb",
				User:         "user",
				Password:     "pass",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantErr: false,
		},
		{
			name: "valid mysql config",
			config: DatabaseConfig{
				Driver:       "mysql",
				Host:         "localhost",
				Port:         3306,
				Name:         "mydb",
				User:         "user",
				Password:     "pass",
				MaxOpenConns: 10,
				MaxIdleConns: 2,
			},
			wantErr: false,
		},
		{
			name: "auto-set defaults for connection pool",
			config: DatabaseConfig{
				Driver:       "postgres",
				Host:         "localhost",
				Port:         5432,
				Name:         "mydb",
				User:         "user",
				Password:     "pass",
				MaxOpenConns: 0, // Should be set to default
				MaxIdleConns: 0, // Should be set to default
			},
			wantErr: false,
		},
		{
			name: "empty driver",
			config: DatabaseConfig{
				Driver:   "",
				Host:     "localhost",
				Port:     5432,
				Name:     "mydb",
				User:     "user",
				Password: "pass",
			},
			wantErr: true,
			errMsg:  "database driver is required",
		},
		{
			name: "invalid driver",
			config: DatabaseConfig{
				Driver:   "mongodb",
				Host:     "localhost",
				Port:     27017,
				Name:     "mydb",
				User:     "user",
				Password: "pass",
			},
			wantErr: true,
			errMsg:  "database driver must be 'postgres' or 'mysql'",
		},
		{
			name: "empty host",
			config: DatabaseConfig{
				Driver:   "postgres",
				Host:     "",
				Port:     5432,
				Name:     "mydb",
				User:     "user",
				Password: "pass",
			},
			wantErr: true,
			errMsg:  "database host is required",
		},
		{
			name: "invalid port - zero",
			config: DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     0,
				Name:     "mydb",
				User:     "user",
				Password: "pass",
			},
			wantErr: true,
			errMsg:  "database port must be positive",
		},
		{
			name: "invalid port - negative",
			config: DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     -1,
				Name:     "mydb",
				User:     "user",
				Password: "pass",
			},
			wantErr: true,
			errMsg:  "database port must be positive",
		},
		{
			name: "empty database name",
			config: DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Name:     "",
				User:     "user",
				Password: "pass",
			},
			wantErr: true,
			errMsg:  "database name is required",
		},
		{
			name: "empty user",
			config: DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Name:     "mydb",
				User:     "",
				Password: "pass",
			},
			wantErr: true,
			errMsg:  "database user is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				// Verify defaults are set
				if tt.config.MaxOpenConns == 0 {
					assert.Equal(t, 25, tt.config.MaxOpenConns)
				}
				if tt.config.MaxIdleConns == 0 {
					assert.Equal(t, 5, tt.config.MaxIdleConns)
				}
			}
		})
	}
}

// TestJWTConfig_Validate tests JWTConfig validation
func TestJWTConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  JWTConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid jwt config",
			config: JWTConfig{
				Secret:          "this-is-a-very-long-secret-key-with-at-least-32-characters",
				ExpirationHours: 24,
			},
			wantErr: false,
		},
		{
			name: "auto-set default expiration",
			config: JWTConfig{
				Secret:          "this-is-a-very-long-secret-key-with-at-least-32-characters",
				ExpirationHours: 0, // Should be set to default
			},
			wantErr: false,
		},
		{
			name: "empty secret",
			config: JWTConfig{
				Secret:          "",
				ExpirationHours: 24,
			},
			wantErr: true,
			errMsg:  "jwt secret is required",
		},
		{
			name: "secret too short",
			config: JWTConfig{
				Secret:          "short-secret",
				ExpirationHours: 24,
			},
			wantErr: true,
			errMsg:  "jwt secret must be at least 32 characters",
		},
		{
			name: "negative expiration hours",
			config: JWTConfig{
				Secret:          "this-is-a-very-long-secret-key-with-at-least-32-characters",
				ExpirationHours: -1,
			},
			wantErr: false, // Should auto-correct to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				// Verify default is set
				if tt.config.ExpirationHours <= 0 {
					assert.Equal(t, 24, tt.config.ExpirationHours)
				}
			}
		})
	}
}

// TestLoggerConfig_Validate tests LoggerConfig validation
func TestLoggerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  LoggerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid logger config - debug json",
			config: LoggerConfig{
				Level:  "debug",
				Format: "json",
			},
			wantErr: false,
		},
		{
			name: "valid logger config - info console",
			config: LoggerConfig{
				Level:  "info",
				Format: "console",
			},
			wantErr: false,
		},
		{
			name: "valid logger config - warn json",
			config: LoggerConfig{
				Level:  "warn",
				Format: "json",
			},
			wantErr: false,
		},
		{
			name: "valid logger config - error json",
			config: LoggerConfig{
				Level:  "error",
				Format: "json",
			},
			wantErr: false,
		},
		{
			name: "case insensitive level - uppercase",
			config: LoggerConfig{
				Level:  "INFO",
				Format: "json",
			},
			wantErr: false,
		},
		{
			name: "case insensitive format - uppercase",
			config: LoggerConfig{
				Level:  "info",
				Format: "JSON",
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: LoggerConfig{
				Level:  "invalid",
				Format: "json",
			},
			wantErr: true,
			errMsg:  "logger level must be one of: debug, info, warn, error",
		},
		{
			name: "invalid log format",
			config: LoggerConfig{
				Level:  "info",
				Format: "xml",
			},
			wantErr: true,
			errMsg:  "logger format must be 'json' or 'console'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestConfig_Validate tests full Config validation
func TestConfig_Validate(t *testing.T) {
	validConfig := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		MasterDatabase: DatabaseConfig{
			Driver:       "postgres",
			Host:         "localhost",
			Port:         5432,
			Name:         "master_db",
			User:         "user",
			Password:     "pass",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
		},
		TenantDatabase: DatabaseConfig{
			Driver:       "postgres",
			Host:         "localhost",
			Port:         5432,
			Name:         "tenant_db",
			User:         "user",
			Password:     "pass",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
		},
		JWT: JWTConfig{
			Secret:          "this-is-a-very-long-secret-key-with-at-least-32-characters",
			ExpirationHours: 24,
		},
		Logger: LoggerConfig{
			Level:  "info",
			Format: "json",
		},
	}

	t.Run("valid full config", func(t *testing.T) {
		err := validConfig.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid server config", func(t *testing.T) {
		cfg := *validConfig
		cfg.Server.Port = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validate server config")
	})

	t.Run("invalid master database config", func(t *testing.T) {
		cfg := *validConfig
		cfg.MasterDatabase.Driver = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validate master database config")
	})

	t.Run("invalid tenant database config", func(t *testing.T) {
		cfg := *validConfig
		cfg.TenantDatabase.Name = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validate tenant database config")
	})

	t.Run("invalid jwt config", func(t *testing.T) {
		cfg := *validConfig
		cfg.JWT.Secret = "short"
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validate jwt config")
	})

	t.Run("invalid logger config", func(t *testing.T) {
		cfg := *validConfig
		cfg.Logger.Level = "invalid"
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validate logger config")
	})
}

// TestLoadConfig tests configuration loading
func TestLoadConfig(t *testing.T) {
	t.Run("load config with environment variables", func(t *testing.T) {
		// Skip this test as environment variable configuration is complex to test
		// The functionality is tested through integration tests
		t.Skip("Environment variable configuration tested through integration tests")
	})

	t.Run("load config with non-existent file", func(t *testing.T) {
		cfg, err := LoadConfig("non_existent_file.yaml")
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "read config file")
	})

	t.Run("load config with invalid data", func(t *testing.T) {
		// Skip this test as environment variable configuration is complex to test
		t.Skip("Environment variable configuration tested through integration tests")
	})
}
