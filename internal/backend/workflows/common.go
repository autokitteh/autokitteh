package workflows

import (
	"fmt"
	"maps"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

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
