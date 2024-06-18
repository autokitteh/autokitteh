package ak

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessiondata"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var CallOptsCtorSymbol = kittehs.Must1(sdktypes.ParseSymbol("callopts"))

var ExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID("ak"))

var TimeoutError = sdktypes.NewSymbolValue(kittehs.Must1(sdktypes.ParseSymbol("timeout")))

type module struct {
	data    *sessiondata.Data
	svcs    *sessionsvcs.Svcs
	syscall sdkexecutor.Function
}

func New(syscall sdkexecutor.Function, data *sessiondata.Data, svcs *sessionsvcs.Svcs) sdkexecutor.Executor {
	mod := &module{data: data, svcs: svcs, syscall: syscall}

	return fixtures.NewBuiltinExecutor(
		ExecutorID,
		sdkmodule.ExportValue("timeout_error", sdkmodule.WithValue(TimeoutError)),
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
		sdkmodule.ExportFunction(
			"is_deployment_active",
			mod.isDeploymentActive,
		),
		sdkmodule.ExportFunction(
			"sleep",
			mod.sleep,
			sdkmodule.WithFlags(
				sdktypes.PureFunctionFlag,        // no need to run in an activity, we must have it in the workflow.
				sdktypes.PrivilidgedFunctionFlag, // provide workflow context.
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

func (m *module) getDeploymentState(ctx context.Context) (sdktypes.DeploymentState, error) {
	did := m.data.Session.DeploymentID()
	if did == sdktypes.InvalidDeploymentID {
		return sdktypes.DeploymentStateUnspecified, nil
	}

	d, err := m.svcs.Deployments.Get(ctx, m.data.Session.DeploymentID())
	if err != nil {
		return sdktypes.DeploymentStateUnspecified, err
	}

	return d.State(), nil
}

func (m *module) isDeploymentActive(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if err := sdkmodule.UnpackArgs(args, kwargs); err != nil {
		return sdktypes.InvalidValue, err
	}

	state, err := m.getDeploymentState(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewBooleanValue(state == sdktypes.DeploymentStateActive), nil
}

func (m *module) sleep(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	args = append([]sdktypes.Value{sdktypes.NewStringValue("sleep")}, args...)
	return m.syscall(ctx, args, kwargs)
}
