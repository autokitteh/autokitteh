package schedule

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	wf "go.autokitteh.dev/autokitteh/internal/backend/workflows"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// TODO: consider design where the workflow is an implementation
// detail of the scheduler and does not need to be exposed via FX

type SchedulerWorkflow struct {
	wf.Workflow
}

type scheduleWorkflowOutput struct{}

func NewSchedulerWorkflow(z *zap.Logger, services wf.Services, tc temporalclient.Client) SchedulerWorkflow {
	return SchedulerWorkflow{wf.Workflow{Z: z, Services: services, Tmprl: tc}}
}

func (swf *SchedulerWorkflow) Start(context.Context) error {
	w := worker.New(swf.Tmprl.Temporal(), wf.ScheduleTaskQueueName, worker.Options{Identity: wf.SchedulerWorkerID})
	w.RegisterWorkflowWithOptions(swf.scheduleWorkflow, workflow.RegisterOptions{Name: wf.SchedulerWorkflow})

	if err := w.Start(); err != nil {
		return fmt.Errorf("schedule wf: worker start: %w", err)
	}
	swf.Z.Info("registered workflow and worker", zap.String("workflow", wf.SchedulerWorkflow), zap.String("worker", wf.SchedulerWorkerID))
	return nil
}

func (swf *SchedulerWorkflow) getSessionData(ctx context.Context, triggerID sdktypes.TriggerID) ([]wf.SessionData, error) {
	z := swf.Z.With(zap.String("trigger_id", triggerID.String()))

	if !triggerID.IsValid() {
		z.Debug("invalid trigger")
		return nil, fmt.Errorf("trigger: %w", sdkerrors.ErrNotFound)
	}

	trigger, err := swf.Services.Triggers.Get(ctx, triggerID)
	if err != nil {
		return nil, fmt.Errorf("get trigger: %w", err)
	}

	if trigger.EventType() != sdktypes.SchedulerEventTriggerType { // sanity
		z.Error("unsupported trigger type, expected `scheduler`", zap.String("trigger_type", trigger.EventType()))
		return nil, fmt.Errorf("unexpected trigger type. Need `scheduler`, got %q", trigger.EventType())
	}

	envID := trigger.EnvID()
	if !envID.IsValid() {
		z.Debug("invalid environment")
		return nil, fmt.Errorf("environment: %w", sdkerrors.ErrNotFound)
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

func (swf *SchedulerWorkflow) newScheduleTickEvent(ctx context.Context, triggerID sdktypes.TriggerID) sdktypes.Event {
	z := swf.Z.With(zap.String("trigger_id", triggerID.String()))

	event := kittehs.Must1(sdktypes.EventFromProto(
		&sdktypes.EventPB{
			EventType:    sdktypes.SchedulerTickEventType,
			Memo:         map[string]string{"trigger_id": triggerID.String()},
			ConnectionId: sdktypes.BuiltinSchedulerConnectionID.String(),
		}))

	eventID, err := swf.Services.Events.Save(ctx, event) // create event
	if err != nil {
		z.Panic("failed creating schedule tick event")
	}

	// REVIEW: re-fetching newly created event is costly (db). Right now we are recreating event without CreatedAt.
	// other solution is to cause events.SaveEvent to return created event and not just ID
	event = kittehs.Must1(sdktypes.EventFromProto(&sdktypes.EventPB{EventId: eventID.String(), EventType: sdktypes.SchedulerTickEventType}))

	swf.CreateEventRecord(ctx, eventID, sdktypes.EventStateProcessing) // add `processing` event record
	return event
}

func (swf *SchedulerWorkflow) scheduleWorkflowInternal(wfctx workflow.Context, ctx context.Context, triggerID sdktypes.TriggerID, tickEvent *sdktypes.Event) error {
	z := swf.Z.With(zap.String("trigger_id", triggerID.String()))
	z.Info("started")

	sds, err := swf.getSessionData(ctx, triggerID)
	if err != nil {
		z.Error("get session data", zap.Error(err))
		return err
	}

	// REVIEW: simplify/extract schedule's version of common:CreateSessionsForWorkflow or keep unified?
	if err = swf.StartSessions(wfctx, *tickEvent, sds); err != nil {
		z.Error("failed start session", zap.Error(err))
		return fmt.Errorf("faield start session: %w", err)
	}

	z.Info("finished")
	return err
}

func (swf *SchedulerWorkflow) scheduleWorkflow(wfctx workflow.Context, triggerID sdktypes.TriggerID) (*scheduleWorkflowOutput, error) {
	var err error

	logger := workflow.GetLogger(wfctx)
	logger.Info("started scheduler workflow", "trigger_id", triggerID.String())

	ctx := temporalclient.NewWorkflowContextAsGOContext(wfctx)
	tickEvent := swf.newScheduleTickEvent(ctx, triggerID) // create tick event and add <processing> state

	state := sdktypes.EventStateCompleted
	if err = swf.scheduleWorkflowInternal(wfctx, ctx, triggerID, &tickEvent); err != nil {
		state = sdktypes.EventStateFailed
		logger.Error("error while scheduler workflow", "trigger_id", triggerID.String(), err)
	} else {
		logger.Info("finished scheduler workflow", "trigger_id", triggerID.String())
	}
	swf.CreateEventRecord(ctx, tickEvent.ID(), state) // add <comeplete|failed> state

	if err != nil {
		err = fmt.Errorf("scheduler wf: %w", err)
	}
	return nil, err
}
