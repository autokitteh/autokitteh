package sdktypes

import (
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionStateError struct {
	object[*SessionStateErrorPB, SessionStateErrorTraits]
}

var InvalidSessionStateError SessionStateError

type SessionStateErrorPB = sessionv1.SessionState_Error

type SessionStateErrorTraits struct{}

func (SessionStateError) isConcreteSessionState() {}

func (SessionStateErrorTraits) Validate(m *SessionStateErrorPB) error {
	return objectField[ProgramError]("error", m.Error)
}

func (SessionStateErrorTraits) StrictValidate(m *SessionStateErrorPB) error {
	return mandatory("error", m.Error)
}

func SessionStateErrorFromProto(m *SessionStateErrorPB) (SessionStateError, error) {
	return FromProto[SessionStateError](m)
}

func StrictSessionStateErrorFromProto(m *SessionStateErrorPB) (SessionStateError, error) {
	return Strict(SessionStateErrorFromProto(m))
}

func (s SessionState) GetError() SessionStateError {
	return forceFromProto[SessionStateError](s.read().Error)
}

func NewSessionStateError(err error, prints []string) SessionState {
	return forceFromProto[SessionState](&SessionStatePB{
		Error: &SessionStateErrorPB{
			Prints: prints,
			Error:  WrapError(err).ToProto(),
		},
	})
}
