package module

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"myapp/internal/pkg/database"
	"myapp/internal/service/master/handler"
	"myapp/internal/service/master/migration"
	"myapp/internal/service/master/repository"
	"myapp/internal/service/master/service"
)

// Module exports master service dependencies
var Module = fx.Options(
	fx.Provide(
		// Master repositories (using masterdb)
		repository.NewRepository,
		
		// Master services
		service.NewService,
		
		// Master handlers
		handler.NewHandler,
	),
	
	// Register migrations
	fx.Invoke(RegisterMigrations),
)

// RegisterMigrations registers database migrations for master service
func RegisterMigrations(dbManager *database.DatabaseManager, logger *zap.Logger) {
	// Use master database for master service migrations
	if err := migration.RunMigrations(dbManager.MasterDB, logger); err != nil {
		logger.Error("Failed to run master migrations", zap.Error(err))
	}
}
