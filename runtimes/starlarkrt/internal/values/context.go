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

type tlsKeyType string

const tlsKey = tlsKeyType("autokitteh-vctx")

func (c *Context) SetTLS(th *starlark.Thread) { th.SetLocal(string(tlsKey), c) }

func FromTLS(th *starlark.Thread) *Context {
	ctx, ok := th.Local(string(tlsKey)).(*Context)
	if !ok {
		return nil
	}
	return ctx
}
