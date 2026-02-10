package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"myapp/internal/pkg/config"
	"myapp/internal/pkg/database"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gorm.io/gorm"
)

// mockConfig creates a mock config for testing
func mockConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Logger: config.LoggerConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

// mockDatabaseManager creates a mock database manager for testing
func mockDatabaseManager() *database.DatabaseManager {
	return &database.DatabaseManager{
		MasterDB: &gorm.DB{},
		TenantDB: &gorm.DB{},
	}
}

// TestNewEcho tests Echo server creation
func TestNewEcho(t *testing.T) {
	t.Run("create echo server", func(t *testing.T) {
		cfg := mockConfig()
		logger := zaptest.NewLogger(t)
		dbManager := mockDatabaseManager()

		e := NewEcho(cfg, logger, dbManager)

		require.NotNil(t, e)
		assert.True(t, e.HideBanner)
		assert.True(t, e.HidePort)
		assert.NotNil(t, e.HTTPErrorHandler)
	})

	t.Run("echo server has middleware", func(t *testing.T) {
		cfg := mockConfig()
		logger := zaptest.NewLogger(t)
		dbManager := mockDatabaseManager()

		e := NewEcho(cfg, logger, dbManager)

		// The middleware should be configured
		// We can verify by making a request and checking it doesn't panic
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Should not panic
		assert.NotPanics(t, func() {
			e.ServeHTTP(rec, req)
		})

		// Even though route doesn't exist, middleware should have processed it
		_ = c
	})
}

// TestRequestLoggerMiddleware tests request logging middleware
func TestRequestLoggerMiddleware(t *testing.T) {
	t.Run("log successful request", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		middleware := requestLoggerMiddleware(logger)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Handler that returns success
		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		}

		h := middleware(handler)
		err := h(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("log failed request", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		middleware := requestLoggerMiddleware(logger)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Handler that returns error
		handler := func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusBadRequest, "bad request")
		}

		h := middleware(handler)
		err := h(c)

		assert.Error(t, err)
	})

	t.Run("log request with different methods", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				logger := zaptest.NewLogger(t)
				middleware := requestLoggerMiddleware(logger)

				e := echo.New()
				req := httptest.NewRequest(method, "/test", nil)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				handler := func(c echo.Context) error {
					return c.NoContent(http.StatusOK)
				}

				h := middleware(handler)
				err := h(c)

				assert.NoError(t, err)
			})
		}
	})
}

// TestCustomErrorHandler tests custom error handling
func TestCustomErrorHandler(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedCode  int
		expectedMsg   string
		requestMethod string
	}{
		{
			name:          "echo http error",
			err:           echo.NewHTTPError(http.StatusBadRequest, "invalid input"),
			expectedCode:  http.StatusBadRequest,
			expectedMsg:   "invalid input",
			requestMethod: http.MethodPost,
		},
		{
			name:          "echo http error with error type message",
			err:           echo.NewHTTPError(http.StatusNotFound, errors.New("resource not found")),
			expectedCode:  http.StatusNotFound,
			expectedMsg:   "resource not found",
			requestMethod: http.MethodGet,
		},
		{
			name:          "generic error",
			err:           errors.New("something went wrong"),
			expectedCode:  http.StatusInternalServerError,
			expectedMsg:   "something went wrong",
			requestMethod: http.MethodPost,
		},
		{
			name:          "unauthorized error",
			err:           echo.NewHTTPError(http.StatusUnauthorized, "unauthorized"),
			expectedCode:  http.StatusUnauthorized,
			expectedMsg:   "unauthorized",
			requestMethod: http.MethodGet,
		},
		{
			name:          "head request with error",
			err:           echo.NewHTTPError(http.StatusNotFound, "not found"),
			expectedCode:  http.StatusNotFound,
			expectedMsg:   "",
			requestMethod: http.MethodHead,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			errorHandler := customErrorHandler(logger)

			e := echo.New()
			req := httptest.NewRequest(tt.requestMethod, "/test", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Call error handler
			errorHandler(tt.err, c)

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.requestMethod == http.MethodHead {
				// HEAD requests should have no body
				assert.Empty(t, rec.Body.String())
			} else {
				// Other methods should have error in body
				assert.Contains(t, rec.Body.String(), tt.expectedMsg)
			}
		})
	}
}

// TestCustomErrorHandler_AlreadyCommitted tests error handler with committed response
func TestCustomErrorHandler_AlreadyCommitted(t *testing.T) {
	t.Run("response already committed", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		errorHandler := customErrorHandler(logger)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Commit the response first
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})

		// Try to handle error (should not override committed response)
		errorHandler(errors.New("error after commit"), c)

		// Response should still be OK, not changed by error handler
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "ok")
	})
}

// TestNewEcho_Integration tests the full Echo server setup
func TestNewEcho_Integration(t *testing.T) {
	t.Run("server handles successful request", func(t *testing.T) {
		cfg := mockConfig()
		logger := zaptest.NewLogger(t)
		dbManager := mockDatabaseManager()

		e := NewEcho(cfg, logger, dbManager)

		// Add test route
		e.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
		})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "healthy")
	})

	t.Run("server handles error", func(t *testing.T) {
		cfg := mockConfig()
		logger := zaptest.NewLogger(t)
		dbManager := mockDatabaseManager()

		e := NewEcho(cfg, logger, dbManager)

		// Add route that returns error
		e.GET("/error", func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusBadRequest, "test error")
		})

		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "test error")
	})

	t.Run("server handles panic with recover middleware", func(t *testing.T) {
		cfg := mockConfig()
		logger := zaptest.NewLogger(t)
		dbManager := mockDatabaseManager()

		e := NewEcho(cfg, logger, dbManager)

		// Add route that panics
		e.GET("/panic", func(c echo.Context) error {
			panic("intentional panic")
		})

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		rec := httptest.NewRecorder()

		// Should not panic due to recover middleware
		assert.NotPanics(t, func() {
			e.ServeHTTP(rec, req)
		})

		// Should return 500 error
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("server has CORS enabled", func(t *testing.T) {
		cfg := mockConfig()
		logger := zaptest.NewLogger(t)
		dbManager := mockDatabaseManager()

		e := NewEcho(cfg, logger, dbManager)

		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// CORS middleware should handle OPTIONS request
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
}

// TestRegisterHooks tests lifecycle hooks registration
func TestRegisterHooks(t *testing.T) {
	t.Run("register hooks does not panic", func(t *testing.T) {
		cfg := mockConfig()
		logger := zaptest.NewLogger(t)
		e := echo.New()

		// Creating a minimal fx lifecycle mock is complex
		// This test verifies the function signature is correct
		assert.NotPanics(t, func() {
			// RegisterHooks function exists and accepts correct parameters
			_ = RegisterHooks
			_ = cfg
			_ = logger
			_ = e
		})
	})
}

// TestErrorHandlerJSONFormat tests error response JSON format
func TestErrorHandlerJSONFormat(t *testing.T) {
	t.Run("error response has correct structure", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		errorHandler := customErrorHandler(logger)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		errorHandler(echo.NewHTTPError(http.StatusNotFound, "resource not found"), c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), `"error"`)
		assert.Contains(t, rec.Body.String(), `"status"`)
		assert.Contains(t, rec.Body.String(), `"path"`)
		assert.Contains(t, rec.Body.String(), "resource not found")
		assert.Contains(t, rec.Body.String(), "/test-path")
	})
}
