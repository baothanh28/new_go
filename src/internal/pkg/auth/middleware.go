package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// UserContext represents user information extracted from JWT
type UserContext struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// JWTMiddleware creates middleware that validates JWT tokens
func JWTMiddleware(service *Service, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			token := extractToken(c.Request())
			if token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization token")
			}
			
			// Validate token
			claims, err := service.ValidateToken(c.Request().Context(), token)
			if err != nil {
				logger.Debug("Token validation failed",
					zap.Error(err),
					zap.String("path", c.Path()))
				
				switch err.(type) {
				case *ErrTokenExpired:
					return echo.NewHTTPError(http.StatusUnauthorized, "token has expired")
				case *ErrTokenRevoked:
					return echo.NewHTTPError(http.StatusUnauthorized, "token has been revoked")
				case *ErrTokenInvalid:
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
				default:
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
				}
			}
			
			// Create user context
			userCtx := &UserContext{
				UserID: claims.UserID,
				Email:  claims.Email,
				Role:   claims.Role,
			}
			
			// Store user context in Echo context
			c.Set("user", userCtx)
			
			return next(c)
		}
	}
}

// RequireRole creates middleware that requires specific roles
func RequireRole(roles ...string) echo.MiddlewareFunc {
	roleMap := make(map[string]bool)
	for _, role := range roles {
		roleMap[role] = true
	}
	
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, err := GetUserFromContext(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not found in context")
			}
			
			if !roleMap[user.Role] {
				return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("insufficient permissions: required roles: %v", roles))
			}
			
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

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(c echo.Context) (uint, error) {
	user, err := GetUserFromContext(c)
	if err != nil {
		return 0, err
	}
	return user.UserID, nil
}
