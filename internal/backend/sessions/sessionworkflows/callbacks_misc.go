package sessionworkflows

import (
	"context"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessiondata"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (w *sessionWorkflow) start(wctx workflow.Context) func(context.Context, sdktypes.RunID, sdktypes.Symbol, sdktypes.CodeLocation, map[string]sdktypes.Value, map[string]string) (sdktypes.SessionID, error) {
	return func(ctx context.Context, rid sdktypes.RunID, project sdktypes.Symbol, loc sdktypes.CodeLocation, inputs map[string]sdktypes.Value, memo map[string]string) (sdktypes.SessionID, error) {
		l := w.l.With(zap.Any("rid", rid), zap.Any("loc", loc), zap.Any("inputs", inputs), zap.Any("memo", memo), zap.Any("project", project))

		l.Info("child session start requested")

		data := w.data

		if !project.IsValid() {
			data.Session = sdktypes.NewSession(data.Build.ID(), loc, inputs, memo).
				WithParentSessionID(data.Session.ID()).
				WithDeploymentID(data.Session.DeploymentID()).
				WithProjectID(data.Session.ProjectID()).
				WithNewID()

			if err := workflow.ExecuteActivity(wctx, createSessionActivityName, data.Session).Get(wctx, nil); err != nil {
				return sdktypes.InvalidSessionID, err
			}
		} else {
			params := createSessionInProjectActivityParams{
				ParentSessionID: data.Session.ID(),
				OrgID:           data.OrgID,
				Project:         project,
				Loc:             loc,
				Inputs:          inputs,
				Memo:            memo,
			}

			// Use Get on a clean sessiondata.Data struct to avoid it following pointers and changing underlying data.
			var adata sessiondata.Data
			if err := workflow.ExecuteActivity(wctx, createSessionInProjectActivityName, params).Get(wctx, &adata); err != nil {
				return sdktypes.InvalidSessionID, err
			}

			data = adata
		}

		f, err := w.ws.StartChildWorkflow(wctx, data)
		if err != nil {
			return sdktypes.InvalidSessionID, err
		}

		sid := data.Session.ID()

		w.children[sid] = f

		w.l.Info("child session started", zap.Any("child", sid), zap.Any("parent", w.data.Session.ID()))

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
