package server

import (
	"go.uber.org/fx"
)

// Module exports server dependency
var Module = fx.Options(
	fx.Provide(NewEcho),
	fx.Invoke(RegisterHooks),
)
