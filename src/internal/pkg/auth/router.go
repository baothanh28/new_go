package auth

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers authentication routes
func RegisterRoutes(e *echo.Echo, handler *Handler, middleware echo.MiddlewareFunc) {
	auth := e.Group("/auth")
	
	// Public routes (no authentication required)
	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Login)
	auth.POST("/refresh", handler.RefreshToken)
	
	// Protected routes (require authentication)
	auth.POST("/logout", handler.Logout, middleware)
	auth.GET("/me", handler.GetCurrentUser, middleware)
}
