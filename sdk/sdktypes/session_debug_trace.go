package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionDebugTrace struct {
	object[*SessionDebugTracePB, SessionDebugTraceTraits]
}

var InvalidSessionDebugTrace SessionDebugTrace

type SessionDebugTracePB = sessionv1.SessionLogRecord_DebugTrace

type SessionDebugTraceTraits struct{}

func (SessionDebugTraceTraits) Validate(m *SessionDebugTracePB) error {
	return errors.Join(
		mandatorySlice("callstack", m.Callstack),
	)
}

func (SessionDebugTraceTraits) StrictValidate(m *SessionDebugTracePB) error {
	return nil
}

func SessionDebugTraceFromProto(m *SessionDebugTracePB) (SessionDebugTrace, error) {
	return FromProto[SessionDebugTrace](m)
}

func StrictSessionDebugTraceFromProto(m *SessionDebugTracePB) (SessionDebugTrace, error) {
	return Strict(SessionDebugTraceFromProto(m))
}

func NewSessionDebugTrace(callstack []CallFrame, extra map[string]string) SessionDebugTrace {
	return forceFromProto[SessionDebugTrace](&SessionDebugTracePB{
		Callstack: kittehs.Transform(callstack, ToProto),
		Extra:     extra,
	})
}
