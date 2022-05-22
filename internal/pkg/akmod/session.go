package akmod

import (
	"context"
	"fmt"
	"time"

	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"

	"github.com/autokitteh/autokitteh/sdk/api/apievent"
	"github.com/autokitteh/autokitteh/sdk/api/apieventsrc"
	"github.com/autokitteh/autokitteh/sdk/api/apilang"
	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	"github.com/autokitteh/L"
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

func NewEventSignal(bindingName string, event *apievent.Event) *sessionEventSignal {
	return &sessionEventSignal{
		// [# signal-event-name #]
		Name:  fmt.Sprintf("%s.%s", bindingName, event.Type()),
		Event: event,
	}
}
