package router

import (
	"myapp/internal/service/product/handler"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RegisterProductTestOnlyRoutes registers all product test only-related routes
func RegisterProductTestOnlyRoutes(
	e *echo.Echo,
	productTestOnlyHandler *handler.ProductTestOnlyHandler,
	logger *zap.Logger,
) {
	logger.Info("Registering product test only routes")

	api := e.Group("/api")

	// Public routes - basic rate limiting
	publicRoutes := api.Group("/product-test-only")
	publicRoutes.GET("", productTestOnlyHandler.GetAllProductTestOnly)
	publicRoutes.GET("/:id", productTestOnlyHandler.GetProductTestOnly)
	publicRoutes.GET("/code/:code", productTestOnlyHandler.GetProductTestOnlyByCode)
	publicRoutes.GET("/type/:type", productTestOnlyHandler.GetProductTestOnlyByType)
	publicRoutes.GET("/search", productTestOnlyHandler.SearchProductTestOnly)

	// Protected routes - authentication required
	// Note: uncomment and configure authMiddleware when authentication is implemented
	protectedRoutes := api.Group("/product-test-only") // , authMiddleware())
	protectedRoutes.POST("", productTestOnlyHandler.CreateProductTestOnly)
	protectedRoutes.PUT("/:id", productTestOnlyHandler.UpdateProductTestOnly)
	protectedRoutes.DELETE("/:id", productTestOnlyHandler.DeleteProductTestOnly)

	logger.Info("Product test only routes registered successfully")
}
