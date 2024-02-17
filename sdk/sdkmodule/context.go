package sdkmodule

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var ctxKey = struct{}{}

type callContext struct {
	module *module
	fnv    sdktypes.Value
}

func wrapCallContext(ctx context.Context, m *module, fnv sdktypes.Value) context.Context {
	return context.WithValue(ctx, ctxKey, &callContext{
		module: m,
		fnv:    fnv,
	})
}

func callContextFromContext(ctx context.Context) *callContext {
	if c, ok := ctx.Value(ctxKey).(*callContext); ok {
		return c
	}
	return &callContext{}
}

func FunctionValueFromContext(ctx context.Context) sdktypes.Value {
	return callContextFromContext(ctx).fnv
}

func FunctionDataFromContext(ctx context.Context) []byte {
	fnv := FunctionValueFromContext(ctx)
	if fnv == nil {
		return nil
	}
	return sdktypes.GetFunctionValueData(fnv)
}
