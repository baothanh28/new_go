package product

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"myapp/internal/service/product/handler"
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
	
	// Product routes (you may want to add authentication middleware here)
	productGroup := api.Group("/products")
	productGroup.GET("", productHandler.ListProducts)
	productGroup.GET("/:id", productHandler.GetProduct)
	productGroup.POST("", productHandler.CreateProduct)
	productGroup.PUT("/:id", productHandler.UpdateProduct)
	productGroup.DELETE("/:id", productHandler.DeleteProduct)
	
	logger.Info("Product routes registered successfully")
}
