package health

import (
	"go.uber.org/fx"
	"myapp/internal/pkg/config"
	"myapp/internal/pkg/logger"
	"myapp/internal/pkg/server"
)

// AppModule combines infrastructure and health service modules
var AppModule = fx.Options(
	// Infrastructure modules
	config.Module,
	logger.Module,
	server.Module,
	
	// Health service module (no database needed)
	Module,
	
	// Router registration
	fx.Invoke(RegisterHealthRoutes),
)
