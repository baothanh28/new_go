package module

import (
	"go.uber.org/fx"
	"myapp/internal/pkg/config"
	"myapp/internal/pkg/database"
	"myapp/internal/pkg/logger"
	"myapp/internal/pkg/server"
	"myapp/internal/router"
	"myapp/internal/service/auth"
	"myapp/internal/service/health"
)

// AppModule combines all application modules
var AppModule = fx.Options(
	// Infrastructure modules
	config.Module,
	logger.Module,
	database.Module,
	server.Module,
	
	// Service modules
	auth.Module,   // Has database: Repository → Service → Handler
	health.Module, // No database: Handler only
	
	// Router module (registers all routes)
	router.Module,
)
