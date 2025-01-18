package sdktypes

import (
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionStateCreated struct {
	object[*SessionStateCreatedPB, SessionStateCreatedTraits]
}

func init() { registerObject[SessionStateCreated]() }

func (SessionStateCreated) isConcreteSessionState() {}

var InvalidSessionStateCreated SessionStateCreated

type SessionStateCreatedPB = sessionv1.SessionState_Created

type SessionStateCreatedTraits struct{ immutableObjectTrait }

func (SessionStateCreatedTraits) Validate(m *SessionStateCreatedPB) error       { return nil }
func (SessionStateCreatedTraits) StrictValidate(m *SessionStateCreatedPB) error { return nil }

func SessionStateCreatedFromProto(m *SessionStateCreatedPB) (SessionStateCreated, error) {
	return FromProto[SessionStateCreated](m)
}

func StrictSessionStateCreatedFromProto(m *SessionStateCreatedPB) (SessionStateCreated, error) {
	return Strict(SessionStateCreatedFromProto(m))
}

func (s SessionState) GetCreated() SessionStateCreated {
	return forceFromProto[SessionStateCreated](s.read().Created)
}

func NewSessionStateCreated() SessionState {
	return forceFromProto[SessionState](&SessionStatePB{
		Created: &SessionStateCreatedPB{},
	})
}
