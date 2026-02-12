package router

import (
	"myapp/internal/service/master/handler"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// RegisterMasterRoutes registers all master-related routes
func RegisterMasterRoutes(
	e *echo.Echo,
	masterHandler *handler.Handler,
	logger *zap.Logger,
) {
	logger.Info("Registering master routes")

	// API routes
	api := e.Group("/api")

	// Health check route
	api.GET("/health", masterHandler.Health)

	// Public group - Rate limited but no authentication
	publicMasters := api.Group("/masters", rateLimitMiddleware())
	publicMasters.GET("", masterHandler.GetMasters)
	publicMasters.GET("/:id", masterHandler.GetMaster)

	// Protected group - Requires authentication + validation
	protectedMasters := api.Group("/masters", authMiddleware(), validateRequestMiddleware())
	protectedMasters.POST("", masterHandler.CreateMaster)
	protectedMasters.PUT("/:id", masterHandler.UpdateMaster)

	// Admin group - Requires authentication + admin role + audit logging
	adminMasters := api.Group("/masters", authMiddleware(), adminOnlyMiddleware(), auditLogMiddleware())
	adminMasters.DELETE("/:id", masterHandler.DeleteMaster)

	logger.Info("Master routes registered successfully")
}

// Example middleware implementations
// Move these to a separate middleware package in production

// rateLimitMiddleware limits the number of requests per IP
func rateLimitMiddleware() echo.MiddlewareFunc {
	return middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)) // 20 requests per second
}

// authMiddleware validates JWT token and sets user context
func authMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Example: Check Authorization header
			// token := c.Request().Header.Get("Authorization")
			// if token == "" {
			//     return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization token")
			// }
			//
			// // Validate token and extract user info
			// user, err := validateToken(token)
			// if err != nil {
			//     return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			// }
			//
			// // Set user in context
			// c.Set("user", user)

			return next(c)
		}
	}
}

// adminOnlyMiddleware checks if user has admin role
func adminOnlyMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Example: Check user role from context
			// user, ok := c.Get("user").(*User)
			// if !ok || user.Role != "admin" {
			//     return echo.NewHTTPError(http.StatusForbidden, "admin access required")
			// }

			return next(c)
		}
	}
}

// validateRequestMiddleware validates request body before processing
func validateRequestMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Example: Pre-validate request body
			// contentType := c.Request().Header.Get("Content-Type")
			// if contentType != "application/json" {
			//     return echo.NewHTTPError(http.StatusBadRequest, "content-type must be application/json")
			// }

			return next(c)
		}
	}
}

// auditLogMiddleware logs all admin actions for compliance
func auditLogMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Example: Log admin action
			// user, _ := c.Get("user").(*User)
			// logger.Info("admin action",
			//     zap.String("user_id", user.ID),
			//     zap.String("action", c.Request().Method),
			//     zap.String("path", c.Request().URL.Path),
			// )

			return next(c)
		}
	}
}
