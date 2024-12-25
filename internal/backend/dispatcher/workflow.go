package dispatcher

import (
	"fmt"
	"maps"

	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/types"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	eventInputsSymbolValue   = sdktypes.NewSymbolValue(sdktypes.NewSymbol("event"))
	triggerInputsSymbolValue = sdktypes.NewSymbolValue(sdktypes.NewSymbol("trigger"))
	dataSymbolValue          = sdktypes.NewSymbolValue(sdktypes.NewSymbol("data"))
)

// Events Workflow
type eventsWorkflowInput struct {
	Event   sdktypes.Event
	Options *sdkservices.DispatchOptions
}

type eventsWorkflowOutput struct {
	Started  []string
	Signaled []string
}

func (d *Dispatcher) startSessions(wctx workflow.Context, event sdktypes.Event, sds []sessionData) ([]sdktypes.SessionID, error) {
	eid := event.ID()

	sl := d.sl.With("event_id", eid)

	// build session inputs.
	// DO NOT PASS Memo. It is not intended for automation use, just auditing.
	eventInputs := event.ToValues()
	eventStruct, err := sdktypes.NewStructValue(eventInputsSymbolValue, eventInputs)
	if err != nil {
		sl.With("err", err).Panicf("could not create event struct: %v", err)
		return nil, err
	}

	inputs := map[string]sdktypes.Value{
		"event": eventStruct,
		"data":  eventInputs["data"],
	}

	var started []sdktypes.SessionID

	for _, sd := range sds {
		sl := sl.With("deployment_id", sd.Deployment.ID(), "trigger_id", sd.Trigger.ID(), "entrypoint", sd.CodeLocation)

		session, err := newSession(event, inputs, sd)
		if err != nil {
			sl.With("err", err).Errorf("could not initialize session: %v", err)
			continue
		}

		var sid sdktypes.SessionID

		if err := workflow.ExecuteActivity(wctx, startSessionActivityName, session).Get(wctx, &sid); err != nil {
			sl.With("err", err).Errorf("session activity: %v", err)
			continue
		}

		sl.With("session_id", sid).Infof("started session %v for %v", sid, eid)

		started = append(started, sid)
	}

	return started, nil
}

func (d *Dispatcher) eventsWorkflow(wctx workflow.Context, input eventsWorkflowInput) (*eventsWorkflowOutput, error) {
	event := input.Event
	eid := event.ID()

	sl := d.sl.With("event_id", eid)
	sl.Infof("events workflow started for %v", eid)

	wctx = temporalclient.WithActivityOptions(wctx, taskQueueName, d.cfg.Activity)

	var sds []sessionData

	if err := workflow.ExecuteActivity(wctx, getEventSessionDataActivityName, event, input.Options).Get(wctx, &sds); err != nil {
		sl.With("err", err).Errorf("could not get session data: %v", err)
		return nil, err
	}

	var started []sdktypes.SessionID

	sl.Infof("found %d dispatch destinations for %v", len(sds), eid)

	if len(sds) != 0 {
		var err error
		if started, err = d.startSessions(wctx, event, sds); err != nil {
			sl.With("err", err).Errorf("could not start sessions for %v: %v", eid, err)
			return nil, err
		}
	}

	// execute waiting signals
	wids, err := d.signalWorkflows(wctx, event)
	if err != nil {
		sl.With("err", err).Errorf("signalling error: %v", err)
	}

	return &eventsWorkflowOutput{
		Started:  kittehs.Transform(started, kittehs.ToString),
		Signaled: wids,
	}, nil
}

func (d *Dispatcher) signalWorkflows(wctx workflow.Context, event sdktypes.Event) ([]string, error) {
	eid := event.ID()

	sl := d.sl.With("event_id", eid, "destination_id", event.DestinationID())

	var signals []*types.Signal
	if err := workflow.ExecuteActivity(wctx, listWaitingSignalsActivityName, event.DestinationID()).Get(wctx, &signals); err != nil {
		return nil, fmt.Errorf("list_waiting_signals: %w", err)
	}

	sl.Debugf("found %d signal candidates for %v", len(signals), eid)

	wg := workflow.NewWaitGroup(wctx)

	var wids []string

	for _, signal := range signals {
		sl := sl.With("signal_id", signal.ID.String(), "workflow_id", signal.WorkflowID, "filter", signal.Filter)

		match, err := event.Matches(signal.Filter)
		if err != nil {
			sl.Infof("invalid signal filter: %v", err)
			continue
		}

		if !match {
			sl.Info("signal filter not matching event, skipping")
			continue
		}

		wids = append(wids, signal.WorkflowID)

		wg.Add(1)

		workflow.Go(wctx, func(wctx workflow.Context) {
			sl := sl.With("workflow_id", signal.WorkflowID, "signal_id", signal.ID)

			err := workflow.ExecuteActivity(
				wctx,
				signalWorkflowActivityName,
				signal.WorkflowID,
				signal.ID,
				eid,
			).Get(wctx, nil)
			if err != nil {
				sl.With("err", err).Errorf("signal workflow %v with %v: %v", signal.WorkflowID, signal.ID, err)
			} else {
				sl.Infof("signaled workflow %v with %v", signal.WorkflowID, signal.ID)
			}

			wg.Done()
		})

		wg.Wait(wctx)
	}

	return wids, nil
}

func newSession(event sdktypes.Event, inputs map[string]sdktypes.Value, data sessionData) (sdktypes.Session, error) {
	memo := make(map[string]string)

	memo["event_id"] = event.ID().String()
	memo["event_uuid"] = event.ID().UUIDValue().String()
	memo["event_type"] = event.Type()
	memo["event_destination_id"] = event.DestinationID().String()
	memo["event_destination_uuid"] = event.DestinationID().UUIDValue().String()

	if t := data.Trigger; t.IsValid() {
		inputs = maps.Clone(inputs)
		triggerInputs := t.ToValues()

		var err error

		if inputs["trigger"], err = sdktypes.NewStructValue(triggerInputsSymbolValue, triggerInputs); err != nil {
			return sdktypes.InvalidSession, fmt.Errorf("trigger: %w", err)
		}

		fs := inputs["data"].GetStruct().Fields()
		maps.Copy(fs, triggerInputs["data"].GetStruct().Fields())
		if inputs["data"], err = sdktypes.NewStructValue(dataSymbolValue, fs); err != nil {
			return sdktypes.InvalidSession, fmt.Errorf("data: %w", err)
		}

		memo["trigger_id"] = t.ID().String()
		memo["trigger_uuid"] = t.ID().UUIDValue().String()
		memo["trigger_source_type"] = t.SourceType().String()
		memo["trigger_name"] = t.Name().String()
	}

	if c := data.Connection; c.IsValid() {
		memo["connection_id"] = c.ID().String()
		memo["connection_uuid"] = c.ID().UUIDValue().String()
		memo["connection_name"] = c.Name().String()
	}

	memo["org_id"] = data.OrgID.String()
	memo["org_uuid"] = data.OrgID.UUIDValue().String()

	pid := data.Deployment.ProjectID()

	memo["project_id"] = pid.String()
	memo["project_uuid"] = pid.UUIDValue().String()

	return sdktypes.NewSession(data.Deployment.BuildID(), data.CodeLocation, inputs, memo).
			WithDeploymentID(data.Deployment.ID()).
			WithEventID(event.ID()).
			WithProjectID(pid),
		nil
}
