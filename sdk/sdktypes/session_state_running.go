package sdktypes

import (
	"errors"

	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionStateRunning struct {
	object[*SessionStateRunningPB, SessionStateRunningTraits]
}

func init() { registerObject[SessionStateRunning]() }

var InvalidSessionStateRunning SessionStateRunning

type SessionStateRunningPB = sessionv1.SessionState_Running

type SessionStateRunningTraits struct{ immutableObjectTrait }

func (SessionStateRunning) isConcreteSessionState() {}

func (SessionStateRunningTraits) Validate(m *SessionStateRunningPB) error {
	return errors.Join(
		idField[RunID]("run_id", m.RunId),
		objectField[Value]("call", m.Call),
	)
}

func (SessionStateRunningTraits) StrictValidate(m *SessionStateRunningPB) error {
	return mandatory("run_id", m.RunId)
}

func SessionStateRunningFromProto(m *SessionStateRunningPB) (SessionStateRunning, error) {
	return FromProto[SessionStateRunning](m)
}

func StrictSessionStateRunningFromProto(m *SessionStateRunningPB) (SessionStateRunning, error) {
	return Strict(SessionStateRunningFromProto(m))
}

func (s SessionStateRunning) Call() Value {
	return forceFromProto[Value](s.read().Call)
}

func (s SessionState) GetRunning() SessionStateRunning {
	return forceFromProto[SessionStateRunning](s.read().Running)
}

func NewSessionStateRunning(rid RunID, callValue Value) SessionState {
	return forceFromProto[SessionState](&sessionv1.SessionState{
		Running: &SessionStateRunningPB{
			RunId: rid.String(),
			Call:  callValue.ToProto(),
		},
	})
}
