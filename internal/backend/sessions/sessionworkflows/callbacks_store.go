package sessionworkflows

import (
	"context"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (w *sessionWorkflow) listStoreValues(wctx workflow.Context) func(context.Context, sdktypes.RunID) ([]string, error) {
	return func(ctx context.Context, _ sdktypes.RunID) ([]string, error) {
		if activity.IsActivity(ctx) {
			return w.ws.listStoreValuesActivity(ctx, w.data.Session.ProjectID())
		}

		var vs []string

		if err := workflow.ExecuteActivity(wctx, listStoreValuesActivityName, w.data.Session.ProjectID()).Get(wctx, &vs); err != nil {
			return nil, err
		}

		return vs, nil
	}
}

func (w *sessionWorkflow) mutateStoreValue(wctx workflow.Context) func(context.Context, sdktypes.RunID, string, string, ...sdktypes.Value) (sdktypes.Value, error) {
	return func(ctx context.Context, _ sdktypes.RunID, key, op string, operands ...sdktypes.Value) (sdktypes.Value, error) {
		if activity.IsActivity(ctx) {
			return w.ws.mutateStoreValueActivity(ctx, w.data.Session.ProjectID(), key, op, operands)
		}

		var v sdktypes.Value

		if err := workflow.ExecuteActivity(wctx, mutateStoreValueActivityName, w.data.Session.ProjectID(), key, op, operands).Get(wctx, &v); err != nil {
			return sdktypes.InvalidValue, err
		}

		return v, nil
	}
}
