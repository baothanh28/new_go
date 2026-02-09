package middleware

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myapp/internal/pkg/database"
)

// RequestContext represents the context of a request (tenant or master)
type RequestContext struct {
	Type     string    // "tenant" or "master"
	TenantID string    // Empty for master requests
	Database *gorm.DB  // Selected database connection
}

// ContextMiddleware creates middleware that determines request context type
// based on HTTP headers and selects appropriate database
func ContextMiddleware(dbManager *database.DatabaseManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID := c.Request().Header.Get("X-Tenant-ID")
			requestType := c.Request().Header.Get("X-Request-Type")
			
			ctx := &RequestContext{}
			
			// Determine request type and select database
			if requestType == "master" || tenantID == "" {
				ctx.Type = "master"
				ctx.Database = dbManager.MasterDB
			} else {
				ctx.Type = "tenant"
				ctx.TenantID = tenantID
				ctx.Database = dbManager.TenantDB
			}
			
			// Store context in Echo context
			c.Set("requestContext", ctx)
			
			return next(c)
		}
	}
}

// GetRequestContext retrieves the RequestContext from Echo context
func GetRequestContext(c echo.Context) (*RequestContext, bool) {
	ctx, ok := c.Get("requestContext").(*RequestContext)
	return ctx, ok
}
