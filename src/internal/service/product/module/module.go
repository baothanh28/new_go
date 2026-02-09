package module

import (
	"github.com/base/src/internal/service/product/handler"
	"github.com/base/src/internal/service/product/repository"
	"github.com/base/src/internal/service/product/service"
	"go.uber.org/fx"
)

// Module exports product service dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewRepository,
		service.NewService,
		handler.NewHandler,
	),
)
