package sessionworkflows

import (
	"context"
	"errors"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (w *sessionWorkflow) newModule() sdkexecutor.Executor {
	flags := sdkmodule.WithFlags(
		sdktypes.PureFunctionFlag,       // no need to run in an activity, we must have it in the workflow.
		sdktypes.PrivilegedFunctionFlag, // provide workflow context.
	)

	return fixtures.NewBuiltinExecutor(
		fixtures.ModuleExecutorID,
		sdkmodule.ExportValue("timeout_error", sdkmodule.WithValue(fixtures.TimeoutError)),
		sdkmodule.ExportFunction("syscall", w.syscall, flags),
		sdkmodule.ExportFunction("sleep", w.sleep, flags),
		sdkmodule.ExportFunction("start", w.start, flags),
		sdkmodule.ExportFunction("subscribe", w.subscribe, flags),
		sdkmodule.ExportFunction("unsubscribe", w.unsubscribe, flags),
		sdkmodule.ExportFunction("next_event", w.nextEvent, flags),
		sdkmodule.ExportFunction("callopts", callopts, flags),
		sdkmodule.ExportFunction("is_deployment_active", w.isDeploymentActive),
	)
}

func callopts(_ context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if len(args) > 0 {
		return sdktypes.InvalidValue, errors.New("expecting only key value arguments")
	}

	return sdktypes.NewStructValue(sdktypes.NewSymbolValue(fixtures.CallOptsCtorSymbol), kwargs)
}

func (w *sessionWorkflow) isDeploymentActive(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if err := sdkmodule.UnpackArgs(args, kwargs); err != nil {
		return sdktypes.InvalidValue, err
	}

	state := sdktypes.DeploymentStateUnspecified

	if did := w.data.Session.DeploymentID(); did.IsValid() {
		d, err := w.ws.svcs.Deployments.Get(authcontext.SetAuthnSystemUser(ctx), did)
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		state = d.State()
	}

	return sdktypes.NewBooleanValue(state == sdktypes.DeploymentStateActive), nil
}
