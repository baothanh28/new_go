package router

import (
	"myapp/internal/service/product/handler"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// RegisterProductRoutes registers all product-related routes
func RegisterProductRoutes(
	e *echo.Echo,
	productHandler *handler.Handler,
	logger *zap.Logger,
) {
	logger.Info("Registering product routes")

	// API routes
	api := e.Group("/api")

	// ==========================================
	// EXAMPLE 1: Route-Level Middleware (per route)
	// ==========================================
	// Middleware applied to individual routes
	/*
	productGroup := api.Group("/products")
	productGroup.GET("", productHandler.GetProducts, rateLimitMiddleware())
	productGroup.GET("/:id", productHandler.GetProduct, rateLimitMiddleware())
	productGroup.POST("", productHandler.CreateProduct, authMiddleware(), validateRequestMiddleware())
	productGroup.PUT("/:id", productHandler.UpdateProduct, authMiddleware(), validateRequestMiddleware())
	productGroup.DELETE("/:id", productHandler.DeleteProduct, authMiddleware(), adminOnlyMiddleware(), auditLogMiddleware())
	*/

	// ==========================================
	// EXAMPLE 2: Group-Level Middleware (RECOMMENDED)
	// ==========================================
	// Middleware applied to entire group - cleaner and more maintainable
	
	// Public group - Rate limited but no authentication
	publicProducts := api.Group("/products", rateLimitMiddleware())
	publicProducts.GET("", productHandler.GetProducts)
	publicProducts.GET("/:id", productHandler.GetProduct)

	// Protected group - Requires authentication + validation
	protectedProducts := api.Group("/products", authMiddleware(), validateRequestMiddleware())
	protectedProducts.POST("", productHandler.CreateProduct)
	protectedProducts.PUT("/:id", productHandler.UpdateProduct)

	// Admin group - Requires authentication + admin role + audit logging
	adminProducts := api.Group("/products", authMiddleware(), adminOnlyMiddleware(), auditLogMiddleware())
	adminProducts.DELETE("/:id", productHandler.DeleteProduct)

	// ==========================================
	// EXAMPLE 3: Nested Groups with Inherited Middleware
	// ==========================================
	/*
	// Base group with common middleware (all routes inherit this)
	products := api.Group("/products", rateLimitMiddleware())

	// Public routes - inherit rate limiting only
	products.GET("", productHandler.GetProducts)
	products.GET("/:id", productHandler.GetProduct)

	// Protected subgroup - inherits rate limiting + adds authentication
	protected := products.Group("", authMiddleware(), validateRequestMiddleware())
	protected.POST("", productHandler.CreateProduct)
	protected.PUT("/:id", productHandler.UpdateProduct)

	// Admin subgroup - inherits all above + adds admin check + audit log
	admin := protected.Group("", adminOnlyMiddleware(), auditLogMiddleware())
	admin.DELETE("/:id", productHandler.DeleteProduct)
	*/

	// ==========================================
	// EXAMPLE 4: Using .Use() to Add Middleware Dynamically
	// ==========================================
	/*
	products := api.Group("/products")
	
	// Add middleware dynamically to the group
	products.Use(rateLimitMiddleware())
	
	// Public routes
	products.GET("", productHandler.GetProducts)
	products.GET("/:id", productHandler.GetProduct)
	
	// Create protected subgroup
	protected := products.Group("")
	protected.Use(authMiddleware())
	protected.Use(validateRequestMiddleware())
	protected.POST("", productHandler.CreateProduct)
	protected.PUT("/:id", productHandler.UpdateProduct)
	
	// Create admin subgroup
	admin := protected.Group("")
	admin.Use(adminOnlyMiddleware())
	admin.Use(auditLogMiddleware())
	admin.DELETE("/:id", productHandler.DeleteProduct)
	*/

	// ==========================================
	// EXAMPLE 5: Mixed Approach (Group + Route-Specific)
	// ==========================================
	/*
	// Group with basic middleware
	products := api.Group("/products", rateLimitMiddleware())
	
	// Public routes - no additional middleware
	products.GET("", productHandler.GetProducts)
	products.GET("/:id", productHandler.GetProduct)
	
	// Protected routes - add auth at route level
	products.POST("", productHandler.CreateProduct, authMiddleware(), validateRequestMiddleware())
	products.PUT("/:id", productHandler.UpdateProduct, authMiddleware(), validateRequestMiddleware())
	
	// Admin route - add both auth and admin at route level
	products.DELETE("/:id", productHandler.DeleteProduct, authMiddleware(), adminOnlyMiddleware(), auditLogMiddleware())
	*/

	logger.Info("Product routes registered successfully")
}

// Example middleware implementations
// Move these to a separate middleware package in production

// rateLimitMiddleware limits the number of requests per IP
func rateLimitMiddleware() echo.MiddlewareFunc {
	return middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)) // 20 requests per second
}

// authMiddleware validates JWT token and sets user context
func authMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Example: Check Authorization header
			// token := c.Request().Header.Get("Authorization")
			// if token == "" {
			//     return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization token")
			// }
			//
			// // Validate token and extract user info
			// user, err := validateToken(token)
			// if err != nil {
			//     return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			// }
			//
			// // Set user in context
			// c.Set("user", user)

			return next(c)
		}
	}
}

// adminOnlyMiddleware checks if user has admin role
func adminOnlyMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Example: Check user role from context
			// user, ok := c.Get("user").(*User)
			// if !ok || user.Role != "admin" {
			//     return echo.NewHTTPError(http.StatusForbidden, "admin access required")
			// }

			return next(c)
		}
	}
}

// validateRequestMiddleware validates request body before processing
func validateRequestMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Example: Pre-validate request body
			// contentType := c.Request().Header.Get("Content-Type")
			// if contentType != "application/json" {
			//     return echo.NewHTTPError(http.StatusBadRequest, "content-type must be application/json")
			// }

			return next(c)
		}
	}
}

// auditLogMiddleware logs all admin actions for compliance
func auditLogMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Example: Log admin action
			// user, _ := c.Get("user").(*User)
			// logger.Info("admin action",
			//     zap.String("user_id", user.ID),
			//     zap.String("action", c.Request().Method),
			//     zap.String("path", c.Request().URL.Path),
			// )

			return next(c)
		}
	}
}
