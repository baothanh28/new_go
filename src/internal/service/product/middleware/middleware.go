package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// AdminOnly middleware ensures only admin users can access the route
func AdminOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user from context (assuming JWT middleware sets this)
			userRole := c.Get("user_role")
			if userRole == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Unauthorized",
				})
			}

			role, ok := userRole.(string)
			if !ok || role != "admin" {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Admin access required",
				})
			}

			return next(c)
		}
	}
}

// ProductExists middleware checks if product exists before processing
func ProductExists() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// This is a placeholder middleware
			// You can implement product existence check here
			// by injecting the service and checking the database
			return next(c)
		}
	}
}

// RateLimitProducts middleware limits product API requests
func RateLimitProducts() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// This is a placeholder for rate limiting
			// You can implement rate limiting logic here
			// using Redis or in-memory store
			return next(c)
		}
	}
}

// ValidateProductOwnership middleware validates user owns the product
func ValidateProductOwnership() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// This is a placeholder middleware
			// Implement ownership validation if products have owners
			return next(c)
		}
	}
}
