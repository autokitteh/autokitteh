package sessionworkflows

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (w *sessionWorkflow) startCallbackSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return telemetry.T().Start(ctx, "sessionWorkflow.callbacks."+name)
}

func (w *sessionWorkflow) start(wctx workflow.Context) func(context.Context, sdktypes.RunID, sdktypes.Symbol, sdktypes.CodeLocation, map[string]sdktypes.Value, map[string]string) (sdktypes.SessionID, error) {
	return func(ctx context.Context, rid sdktypes.RunID, project sdktypes.Symbol, loc sdktypes.CodeLocation, inputs map[string]sdktypes.Value, memo map[string]string) (sdktypes.SessionID, error) {
		inActivity := activity.IsActivity(ctx)

		_, span := w.startCallbackSpan(ctx, "start")
		defer span.End()

		span.SetAttributes(attribute.String("loc", loc.CanonicalString()))

		l := w.l.With(zap.Any("rid", rid), zap.Any("loc", loc), zap.Any("inputs", inputs), zap.Any("memo", memo), zap.Any("project", project))

		l.Info("child session start requested")

		data := w.data
		parentSessionID := data.Session.ID()

		data.Session = sdktypes.NewSession(data.Session.BuildID(), loc, inputs, memo).
			WithParentSessionID(parentSessionID).
			WithDeploymentID(data.Session.DeploymentID()).
			WithProjectID(data.Session.ProjectID()).
			SetDurable(data.Session.IsDurable())

		if project.IsValid() {
			params := getProjectIDAndActiveBuildIDParams{Project: project, OrgID: data.OrgID}

			var (
				resp *getProjectIDAndActiveBuildIDResponse
				err  error
			)

			if inActivity {
				resp, err = w.ws.getProjectIDAndActiveBuildIDActivity(ctx, params)
			} else {
				err = workflow.ExecuteActivity(wctx, getProjectIDAndActiveBuildIDActivityName, &params).Get(wctx, &resp)
			}

			if err != nil {
				return sdktypes.InvalidSessionID, fmt.Errorf("could not get active build ID for project %s: %w", project, err)
			}

			data.Session = data.Session.
				WithProjectID(resp.ProjectID).
				WithBuildID(resp.BuildID)

			l = l.With(zap.Any("child_project_id", resp.ProjectID), zap.Any("child_build_id", resp.BuildID))
		}

		var (
			sid sdktypes.SessionID
			err error
		)

		if inActivity {
			sid, err = w.ws.startChildSessionActivity(ctx, data.Session)
		} else {
			sid, err = w.ws.StartChildWorkflow(wctx, data.Session)
		}

		if err != nil {
			return sdktypes.InvalidSessionID, err
		}

		data.Session = data.Session.WithID(sid)

		l.Info("child session "+sid.String()+" of "+parentSessionID.String()+" started", zap.Any("child_session", data.Session))

		return sid, nil
	}
}

func (w *sessionWorkflow) isDeploymentActive(wctx workflow.Context) func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		_, span := w.startCallbackSpan(ctx, "start")
		defer span.End()

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
