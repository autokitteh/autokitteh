package http

import (
	"context"

	httpint "go.autokitteh.dev/autokitteh/integrations/http"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var ExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID("ihttp"))

func New() sdkexecutor.Executor {
	mod := httpint.NewModule(nil)

	// context.TODO() is supplied here as the http integration does not require
	// use of context in Configure.
	vs := kittehs.Must1(mod.Configure(context.TODO(), ExecutorID, sdktypes.InvalidConnectionID))

	return sdkexecutor.NewExecutor(mod, []sdktypes.ExecutorID{ExecutorID}, vs)
}
