package workflows

import (
	"context"
	"fmt"
	"maps"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Services struct {
	fx.In

	Connections  sdkservices.Connections
	Deployments  sdkservices.Deployments
	Events       sdkservices.Events
	Integrations sdkservices.Integrations
	Projects     sdkservices.Projects
	Triggers     sdkservices.Triggers
	Sessions     sdkservices.Sessions
	Envs         sdkservices.Envs
}

type Workflow struct {
	Z        *zap.Logger
	Services Services
	Tmprl    temporalclient.Client
}

type SessionData struct {
	Deployment            sdktypes.Deployment
	CodeLocation          sdktypes.CodeLocation
	Trigger               sdktypes.Trigger
	AdditionalTriggerData map[string]sdktypes.Value
}

var (
	eventInputsSymbolValue   = sdktypes.NewSymbolValue(kittehs.Must1(sdktypes.ParseSymbol("event")))
	triggerInputsSymbolValue = sdktypes.NewSymbolValue(kittehs.Must1(sdktypes.ParseSymbol("trigger")))
	dataSymbolValue          = sdktypes.NewSymbolValue(kittehs.Must1(sdktypes.ParseSymbol("data")))
)

// used by both dispatcher and scheduler
func CreateSessionsForWorkflow(event sdktypes.Event, sessionsData []SessionData) ([]*sdktypes.Session, error) {
	// DO NOT PASS Memo. It is not intended for automation use, just auditing.
	eventInputs := event.ToValues()
	eventStruct, err := sdktypes.NewStructValue(eventInputsSymbolValue, eventInputs)
	if err != nil {
		return nil, fmt.Errorf("start session: event: %w", err)
	}

	inputs := map[string]sdktypes.Value{
		"event": eventStruct,
		"data":  eventInputs["data"],
	}

	sessions := make([]*sdktypes.Session, len(sessionsData))
	for i, sd := range sessionsData {
		if t := sd.Trigger; t.IsValid() {
			inputs = maps.Clone(inputs)
			triggerInputs := t.ToValues()

			if len(sd.AdditionalTriggerData) != 0 {
				fs := sd.AdditionalTriggerData
				if fs == nil {
					fs = make(map[string]sdktypes.Value)
				}

				if data, ok := triggerInputs["data"]; ok {
					maps.Copy(fs, data.GetStruct().Fields())
				}

				if triggerInputs["data"], err = sdktypes.NewStructValue(dataSymbolValue, fs); err != nil {
					return nil, fmt.Errorf("trigger: %w", err)
				}
			}

			if inputs["trigger"], err = sdktypes.NewStructValue(triggerInputsSymbolValue, triggerInputs); err != nil {
				return nil, fmt.Errorf("trigger: %w", err)
			}

			fs := inputs["data"].GetStruct().Fields()
			maps.Copy(fs, triggerInputs["data"].GetStruct().Fields())
			if inputs["data"], err = sdktypes.NewStructValue(dataSymbolValue, fs); err != nil {
				return nil, fmt.Errorf("data: %w", err)
			}

		}

		dep := sd.Deployment

		session := sdktypes.NewSession(dep.BuildID(), sd.CodeLocation, inputs, nil).
			WithDeploymentID(dep.ID()).
			WithEventID(event.ID()).
			WithEnvID(dep.EnvID())
		sessions[i] = &session
	}
	return sessions, nil
}

func (wf *Workflow) StartSessions(ctx workflow.Context, event sdktypes.Event, sessionsData []SessionData) error {
	sessions, err := CreateSessionsForWorkflow(event, sessionsData)
	if err != nil {
		return fmt.Errorf("schedule wf: start sessions: %w", err)
	}

	goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

	for _, session := range sessions {
		// TODO(ENG-197): change to local activity.
		sessionID, err := wf.Services.Sessions.Start(goCtx, *session)
		if err != nil {
			wf.Z.Panic("could not start session") // Panic in order to make the workflow retry.
		}
		wf.Z.Info("started session", zap.String("session_id", sessionID.String()))
	}
	return nil
}

func (wf *Workflow) CreateEventRecord(ctx context.Context, eventID sdktypes.EventID, state sdktypes.EventState) error {
	record := sdktypes.NewEventRecord(eventID, state)
	if err := wf.Services.Events.AddEventRecord(ctx, record); err != nil {
		wf.Z.Panic("Failed updating event state record", zap.String("eventID", eventID.String()), zap.String("state", state.String()), zap.Error(err))
	}
	return nil
}
