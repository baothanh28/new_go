package auth

import (
	"go.uber.org/fx"
	"myapp/internal/pkg/config"
	"myapp/internal/pkg/database"
	"myapp/internal/pkg/logger"
	"myapp/internal/pkg/server"
)

// AppModule combines infrastructure and auth service modules
var AppModule = fx.Options(
	// Infrastructure modules
	config.Module,
	logger.Module,
	database.Module,
	server.Module,
	
	// Auth service module
	Module,
	
	// Router registration
	fx.Invoke(RegisterAuthRoutes),
)
