package sdktypes

import (
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionStateStopped struct {
	object[*SessionStateStoppedPB, SessionStateStoppedTraits]
}

var InvalidSessionStateStopped SessionStateStopped

type SessionStateStoppedPB = sessionv1.SessionState_Stopped

type SessionStateStoppedTraits struct{}

func (SessionStateStopped) isConcreteSessionState() {}

func (SessionStateStoppedTraits) Validate(m *SessionStateStoppedPB) error {
	return nil
}

func (SessionStateStoppedTraits) StrictValidate(m *SessionStateStoppedPB) error {
	return nil
}

func SessionStateStoppedFromProto(m *SessionStateStoppedPB) (SessionStateStopped, error) {
	return FromProto[SessionStateStopped](m)
}

func StrictSessionStateStoppedFromProto(m *SessionStateStoppedPB) (SessionStateStopped, error) {
	return Strict(SessionStateStoppedFromProto(m))
}

func (s SessionState) GetStopped() SessionStateStopped {
	return forceFromProto[SessionStateStopped](s.read().Stopped)
}

func NewSessionStateStopped(reason string) SessionState {
	return forceFromProto[SessionState](&sessionv1.SessionState{
		Stopped: &SessionStateStoppedPB{
			Reason: reason,
		},
	})
}
