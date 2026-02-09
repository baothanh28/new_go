package auth

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for authentication
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
func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("bind request: %w", err))
	}
	
	// Validate request
	if req.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}
	if req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "password is required")
	}
	if len(req.Password) < 6 {
		return echo.NewHTTPError(http.StatusBadRequest, "password must be at least 6 characters")
	}
	
	user, err := h.service.Register(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("Registration failed", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user": UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Login handles user login
func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("bind request: %w", err))
	}
	
	// Validate request
	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}
	
	resp, err := h.service.Login(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("Login failed", zap.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"token":   resp.Token,
		"user": UserResponse{
			ID:        resp.User.ID,
			Email:     resp.User.Email,
			Role:      resp.User.Role,
			CreatedAt: resp.User.CreatedAt,
		},
	})
}

// Me handles getting current user information
func (h *Handler) Me(c echo.Context) error {
	// Get user from context (set by JWT middleware)
	userCtx, err := GetUserFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found in context")
	}
	
	// Get user details from database
	user, err := h.service.GetUserByID(c.Request().Context(), userCtx.UserID)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user information")
	}
	
	return c.JSON(http.StatusOK, UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	})
}
