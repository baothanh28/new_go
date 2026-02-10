package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"myapp/internal/pkg/database"
)

// mockDatabaseManager creates a mock database manager for testing
func mockDatabaseManager() *database.DatabaseManager {
	return &database.DatabaseManager{
		MasterDB: &gorm.DB{}, // Mock DB
		TenantDB: &gorm.DB{}, // Mock DB
	}
}

// TestContextMiddleware tests the ContextMiddleware function
func TestContextMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		tenantIDHeader string
		requestType    string
		wantType       string
		wantTenantID   string
		wantDatabase   string
	}{
		{
			name:           "master request with explicit type",
			tenantIDHeader: "",
			requestType:    "master",
			wantType:       "master",
			wantTenantID:   "",
			wantDatabase:   "master",
		},
		{
			name:           "master request without tenant ID",
			tenantIDHeader: "",
			requestType:    "",
			wantType:       "master",
			wantTenantID:   "",
			wantDatabase:   "master",
		},
		{
			name:           "tenant request with tenant ID",
			tenantIDHeader: "tenant-123",
			requestType:    "",
			wantType:       "tenant",
			wantTenantID:   "tenant-123",
			wantDatabase:   "tenant",
		},
		{
			name:           "tenant request with UUID",
			tenantIDHeader: "550e8400-e29b-41d4-a716-446655440000",
			requestType:    "",
			wantType:       "tenant",
			wantTenantID:   "550e8400-e29b-41d4-a716-446655440000",
			wantDatabase:   "tenant",
		},
		{
			name:           "master type overrides tenant ID",
			tenantIDHeader: "tenant-123",
			requestType:    "master",
			wantType:       "master",
			wantTenantID:   "",
			wantDatabase:   "master",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.tenantIDHeader != "" {
				req.Header.Set("X-Tenant-ID", tt.tenantIDHeader)
			}
			if tt.requestType != "" {
				req.Header.Set("X-Request-Type", tt.requestType)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			dbManager := mockDatabaseManager()

			// Handler that checks the context
			handler := func(c echo.Context) error {
				ctx, ok := GetRequestContext(c)
				require.True(t, ok)
				require.NotNil(t, ctx)

				// Verify request context
				assert.Equal(t, tt.wantType, ctx.Type)
				assert.Equal(t, tt.wantTenantID, ctx.TenantID)

				// Verify database selection
				if tt.wantDatabase == "master" {
					assert.Equal(t, dbManager.MasterDB, ctx.Database)
				} else {
					assert.Equal(t, dbManager.TenantDB, ctx.Database)
				}

				// Verify tenant ID in Go context
				if tt.wantTenantID != "" {
					tenantID, err := database.GetTenantID(c.Request().Context())
					assert.NoError(t, err)
					assert.Equal(t, tt.wantTenantID, tenantID)
				}

				return c.String(http.StatusOK, "ok")
			}

			// Apply middleware
			middleware := ContextMiddleware(dbManager)
			h := middleware(handler)

			// Execute
			err := h(c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

// TestGetRequestContext tests retrieving RequestContext from Echo context
func TestGetRequestContext(t *testing.T) {
	t.Run("get existing request context", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Set request context
		expectedCtx := &RequestContext{
			Type:     "tenant",
			TenantID: "tenant-123",
			Database: &gorm.DB{},
		}
		c.Set("requestContext", expectedCtx)

		// Get request context
		ctx, ok := GetRequestContext(c)
		assert.True(t, ok)
		assert.Equal(t, expectedCtx, ctx)
	})

	t.Run("get non-existent request context", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Try to get request context without setting it
		ctx, ok := GetRequestContext(c)
		assert.False(t, ok)
		assert.Nil(t, ctx)
	})

	t.Run("get request context with wrong type", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Set wrong type
		c.Set("requestContext", "wrong-type")

		// Try to get request context
		ctx, ok := GetRequestContext(c)
		assert.False(t, ok)
		assert.Nil(t, ctx)
	})
}

// TestContextMiddleware_Integration tests the full middleware flow
func TestContextMiddleware_Integration(t *testing.T) {
	t.Run("full tenant request flow", func(t *testing.T) {
		e := echo.New()
		dbManager := mockDatabaseManager()

		// Apply middleware to the echo instance
		e.Use(ContextMiddleware(dbManager))

		// Add test route
		e.GET("/test", func(c echo.Context) error {
			ctx, ok := GetRequestContext(c)
			if !ok {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "no request context",
				})
			}

			return c.JSON(http.StatusOK, map[string]string{
				"type":      ctx.Type,
				"tenant_id": ctx.TenantID,
			})
		})

		// Create request with tenant ID
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Tenant-ID", "tenant-456")
		rec := httptest.NewRecorder()

		// Serve request
		e.ServeHTTP(rec, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"type":"tenant"`)
		assert.Contains(t, rec.Body.String(), `"tenant_id":"tenant-456"`)
	})

	t.Run("full master request flow", func(t *testing.T) {
		e := echo.New()
		dbManager := mockDatabaseManager()

		// Apply middleware
		e.Use(ContextMiddleware(dbManager))

		// Add test route
		e.GET("/test", func(c echo.Context) error {
			ctx, ok := GetRequestContext(c)
			if !ok {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "no request context",
				})
			}

			return c.JSON(http.StatusOK, map[string]string{
				"type":      ctx.Type,
				"tenant_id": ctx.TenantID,
			})
		})

		// Create master request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-Type", "master")
		rec := httptest.NewRecorder()

		// Serve request
		e.ServeHTTP(rec, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"type":"master"`)
		assert.Contains(t, rec.Body.String(), `"tenant_id":""`)
	})
}

// TestRequestContext_Struct tests the RequestContext struct
func TestRequestContext_Struct(t *testing.T) {
	t.Run("create master request context", func(t *testing.T) {
		ctx := &RequestContext{
			Type:     "master",
			TenantID: "",
			Database: &gorm.DB{},
		}

		assert.Equal(t, "master", ctx.Type)
		assert.Empty(t, ctx.TenantID)
		assert.NotNil(t, ctx.Database)
	})

	t.Run("create tenant request context", func(t *testing.T) {
		ctx := &RequestContext{
			Type:     "tenant",
			TenantID: "tenant-789",
			Database: &gorm.DB{},
		}

		assert.Equal(t, "tenant", ctx.Type)
		assert.Equal(t, "tenant-789", ctx.TenantID)
		assert.NotNil(t, ctx.Database)
	})
}
