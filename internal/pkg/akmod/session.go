package akmod

import (
	"context"
	"time"

	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"

	"github.com/autokitteh/L"
	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apieventsrc"
	"go.autokitteh.dev/sdk/api/apilang"
	"go.autokitteh.dev/sdk/api/apiproject"
	"go.autokitteh.dev/sdk/api/apivalues"
)

var sessionKey struct{}

type Session struct {
	Context        workflow.Context
	UpdateState    func(*apievent.ProjectEventState) error
	SignalChannel  workflow.ReceiveChannel
	L              L.L
	RunSummary     *apilang.RunSummary
	Temporal       temporalclient.Client
	ProjectID      apiproject.ProjectID
	Event          *apievent.Event
	SrcBindingName string
}

func WithSessionContext(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

func getSessionContext(ctx context.Context) *Session {
	return ctx.Value(sessionKey).(*Session)
}

const SessionEventSignalName = "wait_event"

type sessionEventSignal struct {
	Name string

	// For events.
	Event *apievent.Event
}

const (
	syntheticEventSourceID = apieventsrc.EventSourceID("internal.synthetic")
	syntheticEventType     = "signal"
)

func NewSyntheticEvent(origEvent *apievent.Event, name string, v *apivalues.Value) *sessionEventSignal {
	return &sessionEventSignal{
		Name: name,
		Event: apievent.MustNewEvent(
			origEvent.ID(),
			syntheticEventSourceID,
			"",
			name,
			syntheticEventType,
			map[string]*apivalues.Value{"value": v},
			map[string]string{
				"source-event-id":        origEvent.ID().String(),
				"source-event-source-id": origEvent.EventSourceID().String(),
			},
			time.Now(),
		),
	}
}
