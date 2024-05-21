package dispatcher

import (
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Events Workflow
type EventsWorkflowInput struct {
	EventID   sdktypes.EventID
	Options   *sdkservices.DispatchOptions
	TriggerID *sdktypes.TriggerID
}

type eventsWorkflowOutput struct{}
