// +build cgo

package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"myapp/internal/pkg/auth"
)

// setupTestHandler creates a test handler with all dependencies
func setupTestHandler(t *testing.T) (*auth.Handler, *auth.Service) {
	service, _ := setupTestService(t)
	logger := zap.NewNop()
	handler := auth.NewHandler(service, logger)
	return handler, service
}

func TestHandler_Register(t *testing.T) {
	handler, _ := setupTestHandler(t)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(*httptest.ResponseRecorder)
	}{
		{
			name: "valid registration",
			requestBody: RegisterRequest{
				Email:    "register@example.com",
				Password: "SecurePass123",
				Role:     "user",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				var response auth.UserResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "register@example.com", response.Email)
			},
		},
		{
			name: "missing email",
			requestBody: RegisterRequest{
				Password: "SecurePass123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "password too short",
			requestBody: RegisterRequest{
				Email:    "shortpass@example.com",
				Password: "short",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate email",
			requestBody: RegisterRequest{
				Email:    "duplicate@example.com",
				Password: "SecurePass123",
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Register duplicate user first if needed
			if tt.name == "duplicate email" {
				dupReq := RegisterRequest{
					Email:    "duplicate@example.com",
					Password: "SecurePass123",
				}
				_, err := handler.service.Register(context.Background(), &dupReq)
				require.NoError(t, err)
			}

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err = handler.Register(c)
			if tt.expectedStatus != http.StatusOK && tt.expectedStatus != http.StatusCreated {
				assert.Error(t, err)
				httpErr, ok := err.(*echo.HTTPError)
				if ok {
					assert.Equal(t, tt.expectedStatus, httpErr.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(rec)
				}
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	handler, service := setupTestHandler(t)
	ctx := context.Background()

	// Register a user first
		registerReq := &auth.RegisterRequest{
		Email:    "loginhandler@example.com",
		Password: "SecurePass123",
	}
	_, err := service.Register(ctx, registerReq)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    auth.LoginRequest
		expectedStatus int
	}{
		{
			name: "valid login",
			requestBody: LoginRequest{
				Email:    "loginhandler@example.com",
				Password: "SecurePass123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid email",
			requestBody: LoginRequest{
				Email:    "wrong@example.com",
				Password: "SecurePass123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid password",
			requestBody: LoginRequest{
				Email:    "loginhandler@example.com",
				Password: "WrongPassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "missing email",
			requestBody: LoginRequest{
				Password: "SecurePass123",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err = handler.Login(c)
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				var response auth.LoginResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
			} else {
				assert.Error(t, err)
				httpErr, ok := err.(*echo.HTTPError)
				if ok {
					assert.Equal(t, tt.expectedStatus, httpErr.Code)
				}
			}
		})
	}
}

func TestHandler_RefreshToken(t *testing.T) {
	handler, service := setupTestHandler(t)
	ctx := context.Background()

	// Register and login to get refresh token
		registerReq := &auth.RegisterRequest{
		Email:    "refreshhandler@example.com",
		Password: "SecurePass123",
	}
	_, err := service.Register(ctx, registerReq)
	require.NoError(t, err)

		loginReq := &auth.LoginRequest{
		Email:    "refreshhandler@example.com",
		Password: "SecurePass123",
	}
	loginResponse, err := service.Login(ctx, loginReq)
	require.NoError(t, err)

	t.Run("valid refresh", func(t *testing.T) {
		reqBody := auth.RefreshRequest{
			RefreshToken: loginResponse.RefreshToken,
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.RefreshToken(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

				var response auth.RefreshResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.NotEqual(t, loginResponse.RefreshToken, response.RefreshToken) // Should rotate
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		reqBody := auth.RefreshRequest{
			RefreshToken: "invalid_token",
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.RefreshToken(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	})
}

func TestHandler_Logout(t *testing.T) {
	handler, service := setupTestHandler(t)
	ctx := context.Background()

	// Register and login
		registerReq := &auth.RegisterRequest{
		Email:    "logouthandler@example.com",
		Password: "SecurePass123",
	}
	_, err := service.Register(ctx, registerReq)
	require.NoError(t, err)

		loginReq := &auth.LoginRequest{
		Email:    "logouthandler@example.com",
		Password: "SecurePass123",
	}
	loginResponse, err := service.Login(ctx, loginReq)
	require.NoError(t, err)

	t.Run("valid logout", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.Logout(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("logout without token", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.Logout(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	})
}

func TestHandler_GetCurrentUser(t *testing.T) {
	handler, service := setupTestHandler(t)
	ctx := context.Background()

	// Register and login
		registerReq := &auth.RegisterRequest{
		Email:    "mehandler@example.com",
		Password: "SecurePass123",
	}
	_, err := service.Register(ctx, registerReq)
	require.NoError(t, err)

		loginReq := &auth.LoginRequest{
		Email:    "mehandler@example.com",
		Password: "SecurePass123",
	}
	loginResponse, err := service.Login(ctx, loginReq)
	require.NoError(t, err)

	t.Run("get current user", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Apply middleware first
		middleware := auth.JWTMiddleware(service, zap.NewNop())
		err = middleware(func(c echo.Context) error {
			return nil
		})(c)
		require.NoError(t, err)

		err = handler.GetCurrentUser(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

				var response auth.UserResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "mehandler@example.com", response.Email)
	})

	t.Run("get current user without auth", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.GetCurrentUser(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	})
}
