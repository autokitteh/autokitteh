package ak

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var CallOptsCtorSymbol = kittehs.Must1(sdktypes.ParseSymbol("callopts"))

var ExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID("ak"))

func New(syscall sdkexecutor.Function) sdkexecutor.Executor {
	return fixtures.NewBuiltinExecutor(
		ExecutorID,
		sdkmodule.ExportFunction(
			"syscall",
			syscall,
			sdkmodule.WithFlags(
				sdktypes.PureFunctionFlag,           // no need to run in an activity, we must have it in the workflow.
				sdktypes.PrivilidgedFunctionFlag,    // provide workflow context.
				sdktypes.DisablePollingFunctionFlag, // no polling.
			),
		),
		sdkmodule.ExportFunction(
			"callopts",
			callopts,
			sdkmodule.WithFlags(
				sdktypes.PureFunctionFlag,           // no need to run in an activity, we must have it in the workflow.
				sdktypes.DisablePollingFunctionFlag, // no polling.
			),
		),
	)
}

func callopts(_ context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if len(args) > 0 {
		return sdktypes.InvalidValue, fmt.Errorf("expecting only key value arguments")
	}

	return sdktypes.NewStructValue(sdktypes.NewSymbolValue(CallOptsCtorSymbol), kwargs)
}
