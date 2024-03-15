package tls

import (
	"context"

	"go.starlark.net/starlark"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Context struct {
	GoCtx     context.Context
	RunID     sdktypes.RunID
	Callbacks *sdkservices.RunCallbacks
	Globals   starlark.StringDict
}

type tlsKeyType string

const tlsKey = tlsKeyType("autokitteh")

func Set(th *starlark.Thread, ctx *Context) {
	th.SetLocal(string(tlsKey), ctx)
}

func Get(th *starlark.Thread) *Context {
	if ctx, ok := th.Local(string(tlsKey)).(*Context); ok {
		return ctx
	}
	return nil
}
