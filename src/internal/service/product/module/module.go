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
		// Product repositories
		repository.NewRepository,
		repository.NewProductTestOnlyRepository,
		
		// Product services
		service.NewService,
		service.NewProductTestOnlyService,
		
		// Product handlers
		handler.NewHandler,
		handler.NewProductTestOnlyHandler,
	),
)
