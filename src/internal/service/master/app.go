package master

import (
	"go.uber.org/fx"
	"myapp/internal/pkg/config"
	"myapp/internal/pkg/database"
	"myapp/internal/pkg/logger"
	"myapp/internal/pkg/server"
	authmodule "myapp/internal/pkg/auth"
	mastermodule "myapp/internal/service/master/module"
	masterrouter "myapp/internal/service/master/router"
)

// AppModule combines infrastructure and master service modules
var AppModule = fx.Options(
	// Infrastructure modules
	config.Module,
	logger.Module,
	database.Module,
	server.Module,
	
	// Auth module (included in master service)
	authmodule.Module,
	
	// Master service module
	mastermodule.Module,
	
	// Router registration
	fx.Invoke(masterrouter.RegisterMasterRoutes),
)
