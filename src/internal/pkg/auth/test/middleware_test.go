// +build cgo

package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"myapp/internal/pkg/auth"
)

// setupTestMiddleware creates a test middleware with a mock service
func setupTestMiddleware(t *testing.T) (echo.MiddlewareFunc, *auth.Service) {
	service, _ := setupTestService(t)
	logger := zap.NewNop()
	middleware := auth.JWTMiddleware(service, logger)
	return middleware, service
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	middleware, service := setupTestMiddleware(t)
	ctx := context.Background()

	// Register and login to get a token
		registerReq := &auth.RegisterRequest{
		Email:    "middleware@example.com",
		Password: "SecurePass123",
	}
	_, err := service.Register(ctx, registerReq)
	require.NoError(t, err)

		loginReq := &auth.LoginRequest{
		Email:    "middleware@example.com",
		Password: "SecurePass123",
	}
	loginResponse, err := service.Login(ctx, loginReq)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		user, err := auth.GetUserFromContext(c)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "middleware@example.com", user.Email)
		return c.String(http.StatusOK, "OK")
	}

	err = middleware(handler)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	middleware, _ := setupTestMiddleware(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	err := middleware(handler)(c)
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	middleware, _ := setupTestMiddleware(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	err := middleware(handler)(c)
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestJWTMiddleware_InvalidHeaderFormat(t *testing.T) {
	middleware, _ := setupTestMiddleware(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	err := middleware(handler)(c)
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestRequireRole(t *testing.T) {
	_, service := setupTestMiddleware(t)
	ctx := context.Background()

	// Register admin user
		registerReq := &auth.RegisterRequest{
		Email:    "admin@example.com",
		Password: "SecurePass123",
		Role:     "admin",
	}
	_, err := service.Register(ctx, registerReq)
	require.NoError(t, err)

		loginReq := &auth.LoginRequest{
		Email:    "admin@example.com",
		Password: "SecurePass123",
	}
	loginResponse, err := service.Login(ctx, loginReq)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// First apply JWT middleware
	jwtMiddleware := JWTMiddleware(service, zap.NewNop())
	err = jwtMiddleware(func(c echo.Context) error {
		return nil
	})(c)
	require.NoError(t, err)

	// Then apply role middleware
		roleMiddleware := auth.RequireRole("admin")
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	err = roleMiddleware(handler)(c)
	assert.NoError(t, err)
}

func TestRequireRole_InsufficientPermissions(t *testing.T) {
	_, service := setupTestMiddleware(t)
	ctx := context.Background()

	// Register regular user
		registerReq := &auth.RegisterRequest{
		Email:    "user@example.com",
		Password: "SecurePass123",
		Role:     "user",
	}
	_, err := service.Register(ctx, registerReq)
	require.NoError(t, err)

		loginReq := &auth.LoginRequest{
		Email:    "user@example.com",
		Password: "SecurePass123",
	}
	loginResponse, err := service.Login(ctx, loginReq)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// First apply JWT middleware
	jwtMiddleware := JWTMiddleware(service, zap.NewNop())
	err = jwtMiddleware(func(c echo.Context) error {
		return nil
	})(c)
	require.NoError(t, err)

	// Then apply admin role middleware
		roleMiddleware := auth.RequireRole("admin")
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	err = roleMiddleware(handler)(c)
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusForbidden, httpErr.Code)
}

func TestGetUserFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test without user context
	_, err := GetUserFromContext(c)
	assert.Error(t, err)

	// Test with user context
		userCtx := &auth.UserContext{
		UserID: 1,
		Email:  "test@example.com",
		Role:   "user",
	}
	c.Set("user", userCtx)

	user, err := GetUserFromContext(c)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userCtx.UserID, user.UserID)
	assert.Equal(t, userCtx.Email, user.Email)
	assert.Equal(t, userCtx.Role, user.Role)
}

func TestGetUserIDFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test without user context
	_, err := GetUserIDFromContext(c)
	assert.Error(t, err)

	// Test with user context
		userCtx := &auth.UserContext{
		UserID: 123,
		Email:  "test@example.com",
		Role:   "user",
	}
	c.Set("user", userCtx)

	userID, err := GetUserIDFromContext(c)
	assert.NoError(t, err)
	assert.Equal(t, uint(123), userID)
}
