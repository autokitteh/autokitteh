package schedule

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	wf "go.autokitteh.dev/autokitteh/internal/backend/workflows"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type SchedulerWorkflow struct {
	wf.Workflow
}

type scheduleWorkflowInput struct {
	EventID   sdktypes.EventID
	TriggerID sdktypes.TriggerID
}

type scheduleWorkflowOutput struct{}

func NewSchedulerWorkflow(z *zap.Logger, services wf.Services, tc temporalclient.Client) SchedulerWorkflow {
	return SchedulerWorkflow{wf.Workflow{Z: z, Services: services, Tmprl: tc}}
}

func (swf *SchedulerWorkflow) Start(context.Context) error {
	w := worker.New(swf.Tmprl.Temporal(), sdktypes.ScheduleTaskQueueName, worker.Options{Identity: sdktypes.SchedulerWorkerID})
	w.RegisterWorkflowWithOptions(swf.scheduleWorkflow, workflow.RegisterOptions{Name: sdktypes.SchedulerWorkflow})

	if err := w.Start(); err != nil {
		return fmt.Errorf("schedule wf: worker start: %w", err)
	}
	swf.Z.Info("registered workflow and worker", zap.String("workflow", sdktypes.SchedulerWorkflow), zap.String("worker", sdktypes.SchedulerWorkerID))
	return nil
}

func (swf *SchedulerWorkflow) getSessionData(ctx context.Context, triggerID sdktypes.TriggerID) ([]wf.SessionData, error) {
	z := swf.Z.With(zap.String("trigger_id", triggerID.String()))

	if !triggerID.IsValid() {
		z.Debug("invalid trigger")
		return nil, fmt.Errorf("schedule wf: trigger: %w", sdkerrors.ErrNotFound)
	}

	trigger, err := swf.Services.Triggers.Get(ctx, triggerID)
	if err != nil {
		return nil, fmt.Errorf("schedule wf: get trigger: %w", err)
	}

	if trigger.EventType() != sdktypes.SchedulerEventTriggerType { // sanity
		z.Error("shedule wf: unsupported trigger type, expected `scheduler`", zap.String("trigger_type", trigger.EventType()))
		return nil, fmt.Errorf("unexpected trigger type. Need `scheduler`, got %q", trigger.EventType())
	}

	envID := trigger.EnvID()
	if !envID.IsValid() {
		z.Debug("invalid environment")
		return nil, fmt.Errorf("schedule wf: environment: %w", sdkerrors.ErrNotFound)
	}

	depFilter := sdkservices.ListDeploymentsFilter{State: sdktypes.DeploymentStateActive, EnvID: envID}
	deployments, err := swf.Services.Deployments.List(ctx, depFilter)
	if err != nil {
		z.Panic("could not fetch active deployments", zap.Error(err))
	}

	if len(deployments) == 0 {
		z.Debug("no deployments")
		return nil, nil
	}

	cl := trigger.CodeLocation()
	var sds []wf.SessionData
	for _, dep := range deployments {
		sds = append(sds, wf.SessionData{Deployment: dep, CodeLocation: cl, Trigger: trigger})
		z.Debug("deployment found", zap.String("deployment_id", dep.ID().String()))
	}

	return sds, nil
}

func (swf *SchedulerWorkflow) scheduleWorkflow(wfctx workflow.Context, input scheduleWorkflowInput) (*scheduleWorkflowOutput, error) {
	logger := workflow.GetLogger(wfctx)
	logger.Info("started", "event_id", input.EventID)

	z := swf.Z.With(zap.String("event_id", input.EventID.String()))

	event, err := swf.Services.Events.Get(context.TODO(), input.EventID)
	if err != nil {
		return nil, fmt.Errorf("schedule wf: get event: %w", err)
	}

	if !event.IsValid() {
		z.Error("invalid event")
		return nil, sdkerrors.ErrNotFound
	}

	if event.Type() != sdktypes.SchedulerEventTriggerType { // sanity
		z.Error("unexpected event type, expected `scheduler`", zap.String("event_type", event.Type()))
		return nil, fmt.Errorf("unexpected event type. Need `scheduler`, got %q", event.Type())
	}

	ctx := context.Background()

	// Set event to processing
	if err = swf.CreateEventRecord(ctx, input.EventID, sdktypes.EventStateProcessing); err != nil {
		return nil, fmt.Errorf("schedule wf: set event state: %w", err)
	}

	// get event data
	sds, err := swf.getSessionData(ctx, input.TriggerID)
	if err != nil {
		_ = swf.CreateEventRecord(context.Background(), input.EventID, sdktypes.EventStateFailed)
		z.Error("failed to get session data", zap.Error(err))
		logger.Error("shedule wf: failed processing event", "event_id", input.EventID, "error", err)
		return nil, nil
	}

	// start sessions
	if err = swf.StartSessions(wfctx, event, sds); err != nil {
		z.Error("failed start session", zap.Error(err))
		return nil, fmt.Errorf("schedule wf: faield start session: %w", err)
	}

	// set event to Completed
	if err = swf.CreateEventRecord(context.Background(), input.EventID, sdktypes.EventStateCompleted); err != nil {
		err = fmt.Errorf("schedule wf: set event state: %w", err)
	}

	z.Info("finished")
	return nil, err
}
