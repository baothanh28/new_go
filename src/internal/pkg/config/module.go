package config

import (
	"go.uber.org/fx"
)

// Module exports config dependency
var Module = fx.Options(
	fx.Provide(NewConfig),
	fx.Provide(NewAuthConfig), // Provide AuthConfig extracted from Config
)

// NewConfig creates a new Config instance
func NewConfig() (*Config, error) {
	// Try to load config from default path
	cfg, err := LoadConfig("config/config.yaml")
	if err != nil {
		// If default config fails, try without config file (use env vars and defaults)
		cfg, err = LoadConfig("")
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// NewAuthConfig extracts AuthConfig from Config for dependency injection
func NewAuthConfig(cfg *Config) *AuthConfig {
	return &cfg.Auth
}
