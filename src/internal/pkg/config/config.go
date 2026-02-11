package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server         ServerConfig   `mapstructure:"server"`
	MasterDatabase DatabaseConfig `mapstructure:"master_database"`
	TenantDatabase DatabaseConfig `mapstructure:"tenant_database"`
	JWT            JWTConfig      `mapstructure:"jwt"`
	Auth           AuthConfig     `mapstructure:"auth"`
	Logger         LoggerConfig   `mapstructure:"logger"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DatabaseConfig represents database connection configuration
type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Name         string `mapstructure:"name"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret          string `mapstructure:"secret"`
	ExpirationHours int    `mapstructure:"expiration_hours"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	AccessTokenDuration  time.Duration `mapstructure:"access_token_duration"`   // 15 minutes
	RefreshTokenDuration time.Duration `mapstructure:"refresh_token_duration"` // 7 days
	RSAPrivateKeyPath    string        `mapstructure:"rsa_private_key_path"`
	RSAPublicKeyPath     string        `mapstructure:"rsa_public_key_path"`
	Issuer               string        `mapstructure:"issuer"`
	BCryptCost           int           `mapstructure:"bcrypt_cost"`
}

// LoggerConfig represents logger configuration
type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Validate validates the server configuration
func (c *ServerConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("server host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}
	return nil
}

// Validate validates the database configuration
func (c *DatabaseConfig) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("database driver is required")
	}
	if c.Driver != "postgres" && c.Driver != "mysql" {
		return fmt.Errorf("database driver must be 'postgres' or 'mysql', got: %s", c.Driver)
	}
	if c.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Port <= 0 {
		return fmt.Errorf("database port must be positive")
	}
	if c.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = 25 // default value
	}
	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = 5 // default value
	}
	return nil
}

// Validate validates the JWT configuration
func (c *JWTConfig) Validate() error {
	if c.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}
	if len(c.Secret) < 32 {
		return fmt.Errorf("jwt secret must be at least 32 characters")
	}
	if c.ExpirationHours <= 0 {
		c.ExpirationHours = 24 // default value
	}
	return nil
}

// Validate validates the auth configuration
func (c *AuthConfig) Validate() error {
	if c.AccessTokenDuration <= 0 {
		c.AccessTokenDuration = 15 * time.Minute // default: 15 minutes
	}
	if c.RefreshTokenDuration <= 0 {
		c.RefreshTokenDuration = 7 * 24 * time.Hour // default: 7 days
	}
	if c.RSAPrivateKeyPath == "" {
		return fmt.Errorf("rsa_private_key_path is required")
	}
	if c.RSAPublicKeyPath == "" {
		return fmt.Errorf("rsa_public_key_path is required")
	}
	if c.Issuer == "" {
		c.Issuer = "myapp-auth-service" // default issuer
	}
	if c.BCryptCost <= 0 {
		c.BCryptCost = 12 // default bcrypt cost
	}
	if c.BCryptCost < 4 || c.BCryptCost > 31 {
		return fmt.Errorf("bcrypt_cost must be between 4 and 31")
	}
	return nil
}

// Validate validates the logger configuration
func (c *LoggerConfig) Validate() error {
	validLevels := []string{"debug", "info", "warn", "error"}
	level := strings.ToLower(c.Level)
	valid := false
	for _, l := range validLevels {
		if level == l {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("logger level must be one of: debug, info, warn, error")
	}
	
	validFormats := []string{"json", "console"}
	format := strings.ToLower(c.Format)
	valid = false
	for _, f := range validFormats {
		if format == f {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("logger format must be 'json' or 'console'")
	}
	return nil
}

// Validate validates the entire configuration
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("validate server config: %w", err)
	}
	if err := c.MasterDatabase.Validate(); err != nil {
		return fmt.Errorf("validate master database config: %w", err)
	}
	if err := c.TenantDatabase.Validate(); err != nil {
		return fmt.Errorf("validate tenant database config: %w", err)
	}
	if err := c.JWT.Validate(); err != nil {
		return fmt.Errorf("validate jwt config: %w", err)
	}
	if err := c.Auth.Validate(); err != nil {
		return fmt.Errorf("validate auth config: %w", err)
	}
	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("validate logger config: %w", err)
	}
	return nil
}

// LoadConfig loads and validates configuration from file and environment
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	
	// Set defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("jwt.expiration_hours", 24)
	v.SetDefault("auth.access_token_duration", "15m")
	v.SetDefault("auth.refresh_token_duration", "168h") // 7 days
	v.SetDefault("auth.issuer", "myapp-auth-service")
	v.SetDefault("auth.bcrypt_cost", 12)
	
	// Read config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file %s: %w", configPath, err)
		}
	}
	
	// Environment variables support
	v.SetEnvPrefix("MYAPP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	
	// Unmarshal config
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	
	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	
	return cfg, nil
}
