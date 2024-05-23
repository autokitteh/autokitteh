package schedule

import (
	"context"
	"fmt"
	"maps"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type services struct {
	fx.In

	// Connections  sdkservices.Connections
	Deployments sdkservices.Deployments
	Events      sdkservices.Events
	// Integrations sdkservices.Integrations
	// Projects     sdkservices.Projects
	Triggers sdkservices.Triggers
	Sessions sdkservices.Sessions
	// Envs         sdkservices.Envs
}

type SchedulerWorkflow struct {
	tmprl    temporalclient.Client
	services services
	z        *zap.Logger
}

type scheduleWorkflowInput struct {
	EventID   sdktypes.EventID
	TriggerID sdktypes.TriggerID
}

type scheduleWorkflowOutput struct{}

type sessionData struct {
	deployment            sdktypes.Deployment
	codeLocation          sdktypes.CodeLocation
	trigger               sdktypes.Trigger
	additionalTriggerData map[string]sdktypes.Value
}

var (
	eventInputsSymbolValue   = sdktypes.NewSymbolValue(kittehs.Must1(sdktypes.ParseSymbol("event")))
	triggerInputsSymbolValue = sdktypes.NewSymbolValue(kittehs.Must1(sdktypes.ParseSymbol("trigger")))
	dataSymbolValue          = sdktypes.NewSymbolValue(kittehs.Must1(sdktypes.ParseSymbol("data")))
)

func NewSchedulerWorkflow(z *zap.Logger, services services, tc temporalclient.Client) SchedulerWorkflow {
	return SchedulerWorkflow{z: z, services: services, tmprl: tc}
}

func (swf *SchedulerWorkflow) Start(context.Context) error {
	w := worker.New(swf.tmprl.Temporal(), sdktypes.TaskQueueName, worker.Options{})
	w.RegisterWorkflowWithOptions(swf.scheduleWorkflow, workflow.RegisterOptions{Name: sdktypes.SchedulerWorkflow})

	if err := w.Start(); err != nil {
		return fmt.Errorf("scheduler worker start: %w", err)
	}
	swf.z.Info("started scheduler workflow/worker")
	return nil
}

// FIXME: same as dispatchdr. move to common?
func (swf *SchedulerWorkflow) createEventRecord(ctx context.Context, eventID sdktypes.EventID, state sdktypes.EventState) error {
	record := sdktypes.NewEventRecord(eventID, state)
	if err := swf.services.Events.AddEventRecord(ctx, record); err != nil {
		swf.z.Panic("Failed updating event state record", zap.String("eventID", eventID.String()), zap.String("state", state.String()), zap.Error(err))
	}
	return nil
}

// FIXME: same as in dispatcher. Move to common?
func (swf *SchedulerWorkflow) startSessions(ctx workflow.Context, event sdktypes.Event, sessionsData []sessionData) error {
	// DO NOT PASS Memo. It is not intended for automation use, just auditing.
	eventInputs := event.ToValues()
	eventStruct, err := sdktypes.NewStructValue(eventInputsSymbolValue, eventInputs)
	if err != nil {
		return fmt.Errorf("start session: event: %w", err)
	}

	inputs := map[string]sdktypes.Value{
		"event": eventStruct,
		"data":  eventInputs["data"],
	}

	for _, sd := range sessionsData {
		if t := sd.trigger; t.IsValid() {
			inputs = maps.Clone(inputs)
			triggerInputs := t.ToValues()

			if len(sd.additionalTriggerData) != 0 {
				fs := sd.additionalTriggerData
				if fs == nil {
					fs = make(map[string]sdktypes.Value)
				}

				if data, ok := triggerInputs["data"]; ok {
					maps.Copy(fs, data.GetStruct().Fields())
				}

				if triggerInputs["data"], err = sdktypes.NewStructValue(dataSymbolValue, fs); err != nil {
					return fmt.Errorf("trigger: %w", err)
				}
			}

			if inputs["trigger"], err = sdktypes.NewStructValue(triggerInputsSymbolValue, triggerInputs); err != nil {
				return fmt.Errorf("trigger: %w", err)
			}

			fs := inputs["data"].GetStruct().Fields()
			maps.Copy(fs, triggerInputs["data"].GetStruct().Fields())
			if inputs["data"], err = sdktypes.NewStructValue(dataSymbolValue, fs); err != nil {
				return fmt.Errorf("data: %w", err)
			}

		}

		dep := sd.deployment

		session := sdktypes.NewSession(dep.BuildID(), sd.codeLocation, inputs, nil).
			WithDeploymentID(dep.ID()).
			WithEventID(event.ID()).
			WithEnvID(dep.EnvID())

		goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

		// TODO(ENG-197): change to local activity.
		sessionID, err := swf.services.Sessions.Start(goCtx, session)
		if err != nil {
			swf.z.Panic("could not start session") // Panic in order to make the workflow retry.
		}
		swf.z.Info("started session", zap.String("session_id", sessionID.String()))
	}

	return nil
}

func (swf *SchedulerWorkflow) getSessionData(ctx context.Context, triggerID sdktypes.TriggerID) ([]sessionData, error) {
	z := swf.z.With(zap.String("trigger_id", triggerID.String()))

	if !triggerID.IsValid() {
		z.Debug("shedule wf: invalid trigger")
		return nil, fmt.Errorf("schedule wf: trigger: %w", sdkerrors.ErrNotFound)
	}

	trigger, err := swf.services.Triggers.Get(ctx, triggerID)
	if err != nil {
		return nil, fmt.Errorf("schedule wf: get trigger: %w", err)
	}

	if trigger.EventType() != sdktypes.SchedulerEventTriggerType { // sanity
		z.Error("shedule wf: unsupported trigger type, expected `scheduler`", zap.String("trigger_type", trigger.EventType()))
		return nil, fmt.Errorf("unexpected trigger type. Need `scheduler`, got %q", trigger.EventType())
	}

	envID := trigger.EnvID()
	if !envID.IsValid() {
		z.Debug("shedule wf: invalid environment")
		return nil, fmt.Errorf("schedule wf: environment: %w", sdkerrors.ErrNotFound)
	}

	depFilter := sdkservices.ListDeploymentsFilter{State: sdktypes.DeploymentStateActive, EnvID: envID}
	deployments, err := swf.services.Deployments.List(ctx, depFilter)
	if err != nil {
		z.Panic("could not fetch active deployments", zap.Error(err))
	}

	if len(deployments) == 0 {
		z.Debug("schedule wf: no deployments")
		return nil, nil
	}

	cl := trigger.CodeLocation()
	var sds []sessionData
	for _, dep := range deployments {
		sds = append(sds, sessionData{deployment: dep, codeLocation: cl, trigger: trigger})
		z.Debug("deployment found", zap.String("deployment_id", dep.ID().String()))
	}

	return sds, nil
}

// FIXME: rewrite? we don't actually need this schdule workflow and fetch event data all the time. We could prepare it as input once in create/update
// and then other event in cancel/delete
func (swf *SchedulerWorkflow) scheduleWorkflow(wfctx workflow.Context, input scheduleWorkflowInput) (*scheduleWorkflowOutput, error) {
	logger := workflow.GetLogger(wfctx)
	logger.Info("shedule wf: started", "event_id", input.EventID)

	z := swf.z.With(zap.String("event_id", input.EventID.String()))

	event, err := swf.services.Events.Get(context.TODO(), input.EventID)
	if err != nil {
		return nil, fmt.Errorf("schedule wf: get event: %w", err)
	}

	if !event.IsValid() {
		z.Error("shedule wf: invalid event")
		return nil, sdkerrors.ErrNotFound
	}

	if event.Type() != sdktypes.SchedulerEventTriggerType { // sanity
		z.Error("shedule wf: unsupported event type, expected `scheduler`", zap.String("event_type", event.Type()))
		return nil, fmt.Errorf("unexpected event type. Need `scheduler`, got %q", event.Type())
	}

	ctx := context.Background()

	// Set event to processing
	if err = swf.createEventRecord(ctx, input.EventID, sdktypes.EventStateProcessing); err != nil {
		return nil, fmt.Errorf("schedule wf: set event state: %w", err)
	}

	// get event data
	sds, err := swf.getSessionData(ctx, input.TriggerID)
	if err != nil {
		_ = swf.createEventRecord(context.Background(), input.EventID, sdktypes.EventStateFailed)
		z.Error("shedule wf: failed to get session data", zap.Error(err))
		logger.Error("shedule wf: failed processing event", "event_id", input.EventID, "error", err)
		return nil, nil
	}

	// start session
	if err = swf.startSessions(wfctx, event, sds); err != nil {
		z.Error("shedule wf: failed start session", zap.Error(err))
		return nil, fmt.Errorf("schedule wf: faield start session: %w", err)
	}

	// set event to Completed
	if err = swf.createEventRecord(context.Background(), input.EventID, sdktypes.EventStateCompleted); err != nil {
		err = fmt.Errorf("schedule wf: set event state: %w", err)
	}

	z.Info("shedule wf: finished")
	return nil, err
}
