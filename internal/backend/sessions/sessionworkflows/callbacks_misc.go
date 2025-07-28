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
		if activity.IsActivity(ctx) {
			return sdktypes.InvalidSessionID, errForbiddenInActivity
		}

		_, span := w.startCallbackSpan(ctx, "start")
		defer span.End()

		span.SetAttributes(attribute.String("loc", loc.CanonicalString()))

		l := w.l.With(zap.Any("rid", rid), zap.Any("loc", loc), zap.Any("inputs", inputs), zap.Any("memo", memo), zap.Any("project", project))

		l.Info("child session start requested")

		data := w.data

		projectID := data.Session.ProjectID()
		buildID := data.Session.BuildID()

		if project.IsValid() {
			var resp getProjectIDAndActiveBuildIDResponse
			if err := workflow.ExecuteActivity(wctx, getProjectIDAndActiveBuildID, &getProjectIDAndActiveBuildIDParams{Project: project, OrgID: data.OrgID}).Get(wctx, &resp); err != nil {
				return sdktypes.InvalidSessionID, fmt.Errorf("could not get active build ID for project %s: %w", project, err)
			}
			buildID = resp.BuildID
			projectID = resp.ProjectID

		}

		data.Session = sdktypes.NewSession(buildID, loc, inputs, memo).
			WithParentSessionID(data.Session.ID()).
			WithDeploymentID(data.Session.DeploymentID()).
			WithProjectID(projectID)

		sid, err := w.ws.StartChildWorkflow(wctx, data.Session)
		if err != nil {
			return sdktypes.InvalidSessionID, err
		}

		data.Session = data.Session.WithID(sid)

		w.l.Info("child session started", zap.Any("child", sid), zap.Any("parent", w.data.Session.ID()))

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
