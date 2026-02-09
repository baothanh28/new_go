package logger

import (
	"go.uber.org/fx"
)

// Module exports logger dependency
var Module = fx.Options(
	fx.Provide(NewLogger),
)
