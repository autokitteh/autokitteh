package akmod

import (
	"context"

	"go.temporal.io/sdk/workflow"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apievent"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apilang"
	L "gitlab.com/softkitteh/autokitteh/pkg/l"
)

var sessionKey struct{}

type Session struct {
	Context       workflow.Context
	UpdateState   func(*apievent.ProjectEventState) error
	SignalChannel workflow.ReceiveChannel
	L             L.L
	RunSummary    *apilang.RunSummary
}

func WithSessionContext(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

func getSessionContext(ctx context.Context) *Session {
	return ctx.Value(sessionKey).(*Session)
}

const SessionEventSignalName = "wait_event"

type SessionEventSignal struct {
	Event       *apievent.Event
	BindingName string
}
