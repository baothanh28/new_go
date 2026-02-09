package auth

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RegisterAuthRoutes registers all auth-related routes
func RegisterAuthRoutes(
	e *echo.Echo,
	authHandler *Handler,
	jwtMiddleware echo.MiddlewareFunc,
	logger *zap.Logger,
) {
	logger.Info("Registering auth routes")
	
	// API routes
	api := e.Group("/api")
	
	// Public auth routes (no authentication required)
	authGroup := api.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	
	// Protected auth routes (authentication required)
	protectedAuth := api.Group("/auth")
	protectedAuth.Use(jwtMiddleware)
	protectedAuth.GET("/me", authHandler.Me)
	
	logger.Info("Auth routes registered successfully")
}
