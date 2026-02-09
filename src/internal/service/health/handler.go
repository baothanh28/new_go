package health

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Handler handles health check requests
type Handler struct {
	logger    *zap.Logger
	startTime time.Time
}

// NewHandler creates a new health check handler
func NewHandler(logger *zap.Logger) *Handler {
	return &Handler{
		logger:    logger,
		startTime: time.Now(),
	}
}

// Health returns the basic health status of the service
func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "myapp",
		"time":    time.Now().UTC(),
	})
}

// Ready returns the readiness status of the service
func (h *Handler) Ready(c echo.Context) error {
	uptime := time.Since(h.startTime)
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ready",
		"service": "myapp",
		"uptime":  uptime.String(),
		"time":    time.Now().UTC(),
	})
}

// Live returns the liveness status of the service
func (h *Handler) Live(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "alive",
		"time":   time.Now().UTC(),
	})
}
