package module

import (
	"go.uber.org/fx"
	"myapp/internal/service/product/handler"
	"myapp/internal/service/product/repository"
	"myapp/internal/service/product/service"
)

// Module exports product service dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewRepository,
		service.NewService,
		handler.NewHandler,
	),
)
