package database

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module exports database dependency
var Module = fx.Options(
	fx.Provide(NewDatabaseManager),
	fx.Invoke(RegisterHooks),
)

// RegisterHooks registers database lifecycle hooks
func RegisterHooks(lc fx.Lifecycle, dbManager *DatabaseManager, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Database connections established")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing database connections")
			if err := dbManager.Close(); err != nil {
				logger.Error("Failed to close database connections", zap.Error(err))
				return err
			}
			logger.Info("Database connections closed successfully")
			return nil
		},
	})
}
