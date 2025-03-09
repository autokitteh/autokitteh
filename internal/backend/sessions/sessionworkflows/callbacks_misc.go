package sessionworkflows

import (
	"context"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (w *sessionWorkflow) start(wctx workflow.Context) func(context.Context, sdktypes.RunID, sdktypes.CodeLocation, map[string]sdktypes.Value, map[string]string) (string, error) {
	return func(ctx context.Context, rid sdktypes.RunID, loc sdktypes.CodeLocation, inputs map[string]sdktypes.Value, memo map[string]string) (string, error) {
		l := w.l.With(zap.Any("rid", rid), zap.Any("loc", loc), zap.Any("inputs", inputs), zap.Any("memo", memo))

		l.Info("child session start requested")

		session := sdktypes.NewSession(w.data.Build.ID(), loc, inputs, memo).
			WithParentSessionID(w.data.Session.ID()).
			WithDeploymentID(w.data.Session.DeploymentID()).
			WithProjectID(w.data.Session.ProjectID()).
			WithNewID()

		if err := workflow.ExecuteActivity(wctx, createSessionActivityName, session).Get(wctx, nil); err != nil {
			return "", err
		}

		f, err := w.ws.StartChildWorkflow(wctx, session, w.data)
		if err != nil {
			return "", err
		}

		sid := session.ID()

		w.children[sid] = f

		w.l.Info("child session started", zap.Any("child", sid), zap.Any("parent", w.data.Session.ID()))

		return sid.String(), nil
	}
}

func (w *sessionWorkflow) isDeploymentActive(wctx workflow.Context) func(context.Context) (bool, error) {
	return func(context.Context) (bool, error) {
		if did := w.data.Session.DeploymentID(); did.IsValid() {
			var state sdktypes.DeploymentState

			if err := workflow.ExecuteActivity(wctx, getDeploymentStateActivityName, did).Get(wctx, &state); err != nil {
				return false, err
			}

			return state == sdktypes.DeploymentStateActive, nil
		}

		return false, sdkerrors.NewInvalidArgumentError("no deployment associated with session")
	}
}
