package values

import (
	"go.starlark.net/starlark"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Context struct {
	internalFuncs map[string]*starlark.Function
	externalFuncs map[string]sdktypes.Value
	Call          sdkservices.RunCallFunc
	RunID         sdktypes.RunID

	// Used to deterministically set internal function signatures.
	funcSeq uint
}
