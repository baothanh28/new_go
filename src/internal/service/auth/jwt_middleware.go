package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// UserContext represents user information extracted from JWT
type UserContext struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// JWTMiddleware creates middleware that validates JWT tokens
func JWTMiddleware(service *Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractToken(c.Request())
			if token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization token")
			}
			
			claims, err := service.ValidateToken(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("invalid token: %w", err))
			}
			
			// Extract user information from claims
			userID, ok := claims["user_id"].(float64)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}
			
			email, ok := claims["email"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}
			
			role, ok := claims["role"].(string)
			if !ok {
				role = "user" // default role
			}
			
			userCtx := &UserContext{
				UserID: uint(userID),
				Email:  email,
				Role:   role,
			}
			
			c.Set("user", userCtx)
			return next(c)
		}
	}
}

// extractToken extracts the JWT token from Authorization header
func extractToken(req *http.Request) string {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	
	return parts[1]
}

// GetUserFromContext safely extracts user context from Echo context
func GetUserFromContext(c echo.Context) (*UserContext, error) {
	user, ok := c.Get("user").(*UserContext)
	if !ok || user == nil {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}
