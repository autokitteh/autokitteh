package fixtures

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func NewBuiltinIntegrationID(name string) sdktypes.IntegrationID {
	return sdktypes.IntegrationIDFromName(name)
}

func NewBuiltinIntegrationExecutorID(name string) sdktypes.ExecutorID {
	return sdktypes.NewExecutorID(NewBuiltinIntegrationID(name))
}

func NewBuiltinExecutor(name string, opts ...sdkmodule.Optfn) sdkexecutor.Executor {
	mod := sdkmodule.New(opts...)
	vs := kittehs.Must1(mod.Configure(context.TODO(), NewBuiltinIntegrationExecutorID(name), ""))
	return sdkexecutor.NewExecutor(mod, NewBuiltinIntegrationExecutorID(name), vs)
}
