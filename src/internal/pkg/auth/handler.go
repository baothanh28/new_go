package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Handler provides HTTP handlers for authentication endpoints
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new auth handler
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register handles user registration
// POST /api/auth/register
func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	
	// Validate request
	if req.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}
	if req.Password == "" || len(req.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "password must be at least 8 characters")
	}
	
	// Register user
	user, err := h.service.Register(c.Request().Context(), &req)
	if err != nil {
		switch err.(type) {
		case *ErrEmailExists:
			return echo.NewHTTPError(http.StatusConflict, "email already registered")
		default:
			h.logger.Error("Registration failed",
				zap.String("email", req.Email),
				zap.Error(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "registration failed")
		}
	}
	
	return c.JSON(http.StatusCreated, user.ToUserResponse())
}

// Login handles user login
// POST /api/auth/login
func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	
	// Validate request
	if req.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}
	if req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "password is required")
	}
	
	// Login user
	response, err := h.service.Login(c.Request().Context(), &req)
	if err != nil {
		switch err.(type) {
		case *ErrInvalidCredentials:
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		default:
			h.logger.Error("Login failed",
				zap.String("email", req.Email),
				zap.Error(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "login failed")
		}
	}
	
	return c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// POST /api/auth/refresh
func (h *Handler) RefreshToken(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	
	// Validate request
	if req.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh_token is required")
	}
	
	// Refresh token
	response, err := h.service.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		switch err.(type) {
		case *ErrRefreshTokenNotFound:
			return echo.NewHTTPError(http.StatusUnauthorized, "refresh token not found")
		case *ErrTokenExpired:
			return echo.NewHTTPError(http.StatusUnauthorized, "refresh token has expired")
		case *ErrTokenRevoked:
			return echo.NewHTTPError(http.StatusUnauthorized, "refresh token has been revoked")
		default:
			h.logger.Error("Token refresh failed",
				zap.Error(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "token refresh failed")
		}
	}
	
	return c.JSON(http.StatusOK, response)
}

// Logout handles user logout
// POST /api/auth/logout
func (h *Handler) Logout(c echo.Context) error {
	// Extract token from Authorization header
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization token")
	}
	
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
	}
	token := parts[1]
	
	// Logout user
	if err := h.service.Logout(c.Request().Context(), token); err != nil {
		switch err.(type) {
		case *ErrTokenInvalid:
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		default:
			h.logger.Error("Logout failed",
				zap.Error(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "logout failed")
		}
	}
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

// GetCurrentUser returns the current authenticated user
// GET /api/auth/me
func (h *Handler) GetCurrentUser(c echo.Context) error {
	// Get user from context (set by middleware)
	userCtx, err := GetUserFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found in context")
	}
	
	// Get full user details
	user, err := h.service.GetUserByID(c.Request().Context(), userCtx.UserID)
	if err != nil {
		h.logger.Error("Get user failed",
			zap.Uint("user_id", userCtx.UserID),
			zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user")
	}
	
	return c.JSON(http.StatusOK, user.ToUserResponse())
}
