package auth

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

// Module exports auth service dependencies
var Module = fx.Options(
	fx.Provide(
		NewRepository,
		NewService,
		NewHandler,
		NewJWTMiddlewareFunc,
	),
)

// NewJWTMiddlewareFunc provides JWT middleware function for router
func NewJWTMiddlewareFunc(service *Service) echo.MiddlewareFunc {
	return JWTMiddleware(service)
}
