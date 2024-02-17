package dispatcher

import (
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	workflowName  = "event"
	taskQueueName = "events-task-queue"
)

// Events Workflow
type eventsWorkflowInput struct {
	EventID sdktypes.EventID
	Options *sdkservices.DispatchOptions
}

type eventsWorkflowOutput struct{}
