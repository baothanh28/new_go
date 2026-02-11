package uuidv7

import (
	"go.uber.org/fx"
)

// Module exports UUIDv7 generator dependency
var Module = fx.Options(
	fx.Provide(NewGenerator),
)
