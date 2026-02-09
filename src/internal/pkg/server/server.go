package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"myapp/internal/pkg/config"
	"myapp/internal/pkg/database"
	custommw "myapp/internal/pkg/middleware"
)

// NewEcho creates a new Echo server instance
func NewEcho(cfg *config.Config, logger *zap.Logger, dbManager *database.DatabaseManager) *echo.Echo {
	e := echo.New()
	
	// Hide Echo banner
	e.HideBanner = true
	e.HidePort = true
	
	// Configure custom error handler
	e.HTTPErrorHandler = customErrorHandler(logger)
	
	// Global middleware chain (order matters!)
	e.Use(middleware.Recover())
	e.Use(requestLoggerMiddleware(logger))
	e.Use(custommw.ContextMiddleware(dbManager)) // Tenant/Master context detection
	e.Use(middleware.CORS())
	
	return e
}

// requestLoggerMiddleware creates a middleware for request logging
func requestLoggerMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			
			logger.Info("Incoming request",
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.String("remote_addr", req.RemoteAddr),
			)
			
			err := next(c)
			
			res := c.Response()
			logger.Info("Request completed",
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.Int("status", res.Status),
			)
			
			return err
		}
	}
}

// customErrorHandler handles errors and returns appropriate responses
func customErrorHandler(logger *zap.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		message := "Internal server error"
		
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			if msg, ok := he.Message.(string); ok {
				message = msg
			} else if msg, ok := he.Message.(error); ok {
				message = msg.Error()
			}
		} else {
			message = err.Error()
		}
		
		logger.Error("Request error",
			zap.Int("status", code),
			zap.String("message", message),
			zap.String("path", c.Request().URL.Path),
			zap.Error(err),
		)
		
		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead {
				c.NoContent(code)
			} else {
				c.JSON(code, map[string]interface{}{
					"error":   message,
					"status":  code,
					"path":    c.Request().URL.Path,
				})
			}
		}
	}
}

// RegisterHooks registers server lifecycle hooks
func RegisterHooks(lc fx.Lifecycle, e *echo.Echo, cfg *config.Config, logger *zap.Logger) {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting HTTP server", zap.String("addr", addr))
				if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
					logger.Fatal("Server startup failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down HTTP server")
			if err := e.Shutdown(ctx); err != nil {
				logger.Error("Server shutdown failed", zap.Error(err))
				return err
			}
			logger.Info("HTTP server stopped successfully")
			return nil
		},
	})
}
