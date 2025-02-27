package sessionworkflows

import (
	"context"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (w *sessionWorkflow) start(workflow.Context) func(context.Context, sdktypes.RunID, sdktypes.CodeLocation, map[string]sdktypes.Value, map[string]string) (sdktypes.SessionID, error) {
	return func(ctx context.Context, rid sdktypes.RunID, loc sdktypes.CodeLocation, inputs map[string]sdktypes.Value, memo map[string]string) (sdktypes.SessionID, error) {
		l := w.l.With(zap.Any("rid", rid), zap.Any("loc", loc), zap.Any("inputs", inputs), zap.Any("memo", memo))

		l.Info("session start requested")

		session := sdktypes.NewSession(w.data.Build.ID(), loc, inputs, memo).
			WithParentSessionID(w.data.Session.ID()).
			WithDeploymentID(w.data.Session.DeploymentID()).
			WithProjectID(w.data.Session.ProjectID())

		sid, err := w.ws.sessions.Start(authcontext.SetAuthnSystemUser(ctx), session)
		if err != nil {
			return sdktypes.InvalidSessionID, err
		}

		w.l.Info("session started", zap.Any("sid", sid))

		return sid, nil
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
