package auth

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"myapp/internal/pkg/database"
)

// Module exports auth dependency injection module
var Module = fx.Options(
	// Provide dependencies
	fx.Provide(NewTokenManager),
	fx.Provide(NewRepository),
	fx.Provide(NewTokenRepository),
	fx.Provide(NewService),
	fx.Provide(NewHandler),
	
	// Invoke setup functions
	fx.Invoke(RegisterMigrations),
	fx.Invoke(RegisterRoutesWithMiddleware),
	fx.Invoke(StartCleanupWorker),
)

// RegisterMigrations registers database migrations for auth tables
func RegisterMigrations(dbManager *database.DatabaseManager, logger *zap.Logger) {
	db := dbManager.MasterDB
	
	// Auto-migrate auth tables
	if err := db.AutoMigrate(
		&User{},
		&RefreshToken{},
		&TokenBlacklist{},
	); err != nil {
		logger.Error("Failed to migrate auth tables", zap.Error(err))
		return
	}
	
	logger.Info("Auth tables migrated successfully")
}

// RegisterRoutesWithMiddleware registers auth routes with JWT middleware
func RegisterRoutesWithMiddleware(
	e *echo.Echo,
	handler *Handler,
	service *Service,
	logger *zap.Logger,
) {
	middleware := JWTMiddleware(service, logger)
	RegisterRoutes(e, handler, middleware)
}

// StartCleanupWorker starts a background worker to periodically clean up expired tokens
func StartCleanupWorker(
	lc fx.Lifecycle,
	tokenRepo *TokenRepository,
	logger *zap.Logger,
) {
	// Create a context that will be cancelled when the app stops
	workerCtx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start cleanup worker in background
			go func() {
				ticker := time.NewTicker(1 * time.Hour) // Run cleanup every hour
				defer ticker.Stop()
				
				for {
					select {
					case <-ticker.C:
						cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
						if err := tokenRepo.CleanupExpiredTokens(cleanupCtx); err != nil {
							logger.Error("Failed to cleanup expired tokens", zap.Error(err))
						} else {
							logger.Debug("Cleaned up expired tokens")
						}
						cleanupCancel()
					case <-workerCtx.Done():
						logger.Info("Cleanup worker stopped")
						return
					}
				}
			}()
			
			logger.Info("Token cleanup worker started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping token cleanup worker")
			cancel()
			return nil
		},
	})
}
