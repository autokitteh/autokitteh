package ak

import (
	"go.autokitteh.dev/autokitteh/backend/internal/fixtures"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func New(syscall sdkexecutor.Function) sdkexecutor.Executor {
	return fixtures.NewBuiltinExecutor(
		"ak",
		sdkmodule.ExportFunction(
			"syscall",
			syscall,
			sdkmodule.WithFlags(
				sdktypes.PureFunctionFlag,           // no need to run in an activity, we must have it in the workflow.
				sdktypes.PrivilidgedFunctionFlag,    // provide workflow context.
				sdktypes.DisablePollingFunctionFlag, // no polling.
			),
		),
	)
}
