package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"myapp/internal/pkg/config"
)

// TestParseLogLevel tests log level parsing
func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantLevel zapcore.Level
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "debug level",
			level:     "debug",
			wantLevel: zapcore.DebugLevel,
			wantErr:   false,
		},
		{
			name:      "info level",
			level:     "info",
			wantLevel: zapcore.InfoLevel,
			wantErr:   false,
		},
		{
			name:      "warn level",
			level:     "warn",
			wantLevel: zapcore.WarnLevel,
			wantErr:   false,
		},
		{
			name:      "warning level (alias)",
			level:     "warning",
			wantLevel: zapcore.WarnLevel,
			wantErr:   false,
		},
		{
			name:      "error level",
			level:     "error",
			wantLevel: zapcore.ErrorLevel,
			wantErr:   false,
		},
		{
			name:      "case insensitive - uppercase",
			level:     "INFO",
			wantLevel: zapcore.InfoLevel,
			wantErr:   false,
		},
		{
			name:      "case insensitive - mixed case",
			level:     "WaRn",
			wantLevel: zapcore.WarnLevel,
			wantErr:   false,
		},
		{
			name:      "invalid level",
			level:     "invalid",
			wantLevel: zapcore.InfoLevel, // defaults to info on error
			wantErr:   true,
			errMsg:    "unknown log level: invalid",
		},
		{
			name:      "empty level",
			level:     "",
			wantLevel: zapcore.InfoLevel,
			wantErr:   true,
			errMsg:    "unknown log level:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := parseLogLevel(tt.level)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
			
			assert.Equal(t, tt.wantLevel, level)
		})
	}
}

// TestNewLogger tests logger creation
func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "create logger with json format",
			config: &config.Config{
				Logger: config.LoggerConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantErr: false,
		},
		{
			name: "create logger with console format",
			config: &config.Config{
				Logger: config.LoggerConfig{
					Level:  "debug",
					Format: "console",
				},
			},
			wantErr: false,
		},
		{
			name: "create logger with warn level",
			config: &config.Config{
				Logger: config.LoggerConfig{
					Level:  "warn",
					Format: "json",
				},
			},
			wantErr: false,
		},
		{
			name: "create logger with error level",
			config: &config.Config{
				Logger: config.LoggerConfig{
					Level:  "error",
					Format: "console",
				},
			},
			wantErr: false,
		},
		{
			name: "create logger with uppercase format",
			config: &config.Config{
				Logger: config.LoggerConfig{
					Level:  "info",
					Format: "JSON",
				},
			},
			wantErr: false,
		},
		{
			name: "create logger with invalid level",
			config: &config.Config{
				Logger: config.LoggerConfig{
					Level:  "invalid",
					Format: "json",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, logger)
				
				// Verify logger is functional by writing a test log
				logger.Info("test log message")
				
				// Clean up
				logger.Sync()
			}
		})
	}
}

// TestNewLogger_LogLevels tests that logger respects configured log levels
func TestNewLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		format   string
	}{
		{
			name:   "debug level with json",
			level:  "debug",
			format: "json",
		},
		{
			name:   "info level with console",
			level:  "info",
			format: "console",
		},
		{
			name:   "warn level with json",
			level:  "warn",
			format: "json",
		},
		{
			name:   "error level with console",
			level:  "error",
			format: "console",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Logger: config.LoggerConfig{
					Level:  tt.level,
					Format: tt.format,
				},
			}

			logger, err := NewLogger(cfg)
			require.NoError(t, err)
			require.NotNil(t, logger)
			
			// Verify logger can handle all log levels
			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message")
			logger.Error("error message")
			
			// Clean up
			logger.Sync()
		})
	}
}

// TestNewLogger_CaseInsensitivity tests case insensitive format handling
func TestNewLogger_CaseInsensitivity(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{
			name:   "lowercase json",
			format: "json",
		},
		{
			name:   "uppercase JSON",
			format: "JSON",
		},
		{
			name:   "mixed case Json",
			format: "Json",
		},
		{
			name:   "lowercase console",
			format: "console",
		},
		{
			name:   "uppercase CONSOLE",
			format: "CONSOLE",
		},
		{
			name:   "mixed case Console",
			format: "Console",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Logger: config.LoggerConfig{
					Level:  "info",
					Format: tt.format,
				},
			}

			logger, err := NewLogger(cfg)
			require.NoError(t, err)
			require.NotNil(t, logger)
			
			logger.Info("test message")
			logger.Sync()
		})
	}
}
