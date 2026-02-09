package router

import (
	"go.uber.org/fx"
)

// Module exports router dependency
var Module = fx.Options(
	fx.Provide(NewRouter),
	fx.Invoke(func(r *Router) {
		r.RegisterRoutes()
	}),
)
