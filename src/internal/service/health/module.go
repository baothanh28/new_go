package health

import (
	"go.uber.org/fx"
)

// Module exports health service dependencies
var Module = fx.Options(
	fx.Provide(NewHandler),
)
