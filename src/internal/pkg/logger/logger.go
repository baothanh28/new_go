package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"myapp/internal/pkg/config"
)

// NewLogger creates a new zap logger based on configuration
func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	var zapConfig zap.Config
	
	// Set config based on format
	if strings.ToLower(cfg.Logger.Format) == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	
	// Set log level
	level, err := parseLogLevel(cfg.Logger.Level)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	
	// Build logger
	logger, err := zapConfig.Build(
		zap.AddCallerSkip(0),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}
	
	return logger, nil
}

// parseLogLevel converts string log level to zapcore.Level
func parseLogLevel(level string) (zapcore.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}
