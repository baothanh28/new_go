package health

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RegisterHealthRoutes registers all health check routes
func RegisterHealthRoutes(
	e *echo.Echo,
	healthHandler *Handler,
	logger *zap.Logger,
) {
	logger.Info("Registering health routes")
	
	// Health check routes (public, no authentication required)
	e.GET("/health", healthHandler.Health)
	e.GET("/health/ready", healthHandler.Ready)
	e.GET("/health/live", healthHandler.Live)
	
	logger.Info("Health routes registered successfully")
}
