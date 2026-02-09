package module

import (
	"go.uber.org/fx"
	"myapp/internal/pkg/config"
	"myapp/internal/pkg/database"
	"myapp/internal/pkg/logger"
	"myapp/internal/pkg/server"
	"myapp/internal/service/auth"
	"myapp/internal/service/health"
	"myapp/internal/service/product"
	productmodule "myapp/internal/service/product/module"
)

// AppModule combines all application modules and services
// This module includes all services together for a monolithic deployment
var AppModule = fx.Options(
	// Infrastructure modules
	config.Module,
	logger.Module,
	database.Module,
	server.Module,
	
	// Service modules
	auth.Module,         // Has database: Repository → Service → Handler
	health.Module,       // No database: Handler only
	productmodule.Module, // Has database: Repository → Service → Handler
	
	// Route registration for all services
	fx.Invoke(auth.RegisterAuthRoutes),
	fx.Invoke(health.RegisterHealthRoutes),
	fx.Invoke(product.RegisterProductRoutes),
)
