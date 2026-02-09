package product

import (
	"go.uber.org/fx"
	"myapp/internal/pkg/config"
	"myapp/internal/pkg/database"
	"myapp/internal/pkg/logger"
	"myapp/internal/pkg/server"
	productmodule "myapp/internal/service/product/module"
	productrouter "myapp/internal/service/product/router"
)

// AppModule combines infrastructure and product service modules
var AppModule = fx.Options(
	// Infrastructure modules
	config.Module,
	logger.Module,
	database.Module,
	server.Module,
	
	// Product service module
	productmodule.Module,
	
	// Router registration
	fx.Invoke(productrouter.RegisterProductRoutes),
	fx.Invoke(productrouter.RegisterProductTestOnlyRoutes),
)
