package router

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"myapp/internal/service/auth"
	"myapp/internal/service/health"
)

// Router registers all application routes
type Router struct {
	echo          *echo.Echo
	authHandler   *auth.Handler
	healthHandler *health.Handler
	jwtMiddleware echo.MiddlewareFunc
	logger        *zap.Logger
}

// NewRouter creates a new router instance
func NewRouter(
	e *echo.Echo,
	authHandler *auth.Handler,
	healthHandler *health.Handler,
	jwtMiddleware echo.MiddlewareFunc,
	logger *zap.Logger,
) *Router {
	return &Router{
		echo:          e,
		authHandler:   authHandler,
		healthHandler: healthHandler,
		jwtMiddleware: jwtMiddleware,
		logger:        logger,
	}
}

// RegisterRoutes registers all application routes
func (r *Router) RegisterRoutes() {
	r.logger.Info("Registering application routes")
	
	// Health check routes (public, no authentication required)
	r.echo.GET("/health", r.healthHandler.Health)
	r.echo.GET("/health/ready", r.healthHandler.Ready)
	r.echo.GET("/health/live", r.healthHandler.Live)
	
	// API routes
	api := r.echo.Group("/api")
	
	// Public auth routes (no authentication required)
	authGroup := api.Group("/auth")
	authGroup.POST("/register", r.authHandler.Register)
	authGroup.POST("/login", r.authHandler.Login)
	
	// Protected auth routes (authentication required)
	protectedAuth := api.Group("/auth")
	protectedAuth.Use(r.jwtMiddleware)
	protectedAuth.GET("/me", r.authHandler.Me)
	
	r.logger.Info("Routes registered successfully")
}
