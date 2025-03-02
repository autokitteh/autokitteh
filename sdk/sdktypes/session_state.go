package sdktypes

import (
	"errors"

	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionState struct {
	object[*SessionStatePB, SessionStateTraits]
}

func init() { registerObject[SessionState]() }

type SessionStatePB = sessionv1.SessionState

type SessionStateTraits struct{ immutableObjectTrait }

func (SessionStateTraits) Validate(m *SessionStatePB) error {
	return errors.Join(
		objectField[SessionStateStopped]("stopped", m.Stopped),
		objectField[SessionStateCompleted]("completed", m.Completed),
		objectField[SessionStateCreated]("created", m.Created),
		objectField[SessionStateError]("error", m.Error),
		objectField[SessionStateRunning]("running", m.Running),
	)
}

func (SessionStateTraits) StrictValidate(m *SessionStatePB) error { return oneOfMessage(m) }

func (p SessionState) Type() SessionStateType {
	pb := p.read()

	if pb.Completed != nil {
		return SessionStateTypeCompleted
	}

	if pb.Created != nil {
		return SessionStateTypeCreated
	}

	if pb.Error != nil {
		return SessionStateTypeError
	}

	if pb.Running != nil {
		return SessionStateTypeRunning
	}

	if pb.Stopped != nil {
		return SessionStateTypeStopped
	}

	return SessionStateTypeUnspecified
}

type concreteSessionState interface {
	Object

	isConcreteSessionState()
}

func (p SessionState) Concrete() concreteSessionState {
	switch p.Type() {
	case SessionStateTypeCompleted:
		return p.GetCompleted()
	case SessionStateTypeCreated:
		return p.GetCreated()
	case SessionStateTypeError:
		return p.GetError()
	case SessionStateTypeRunning:
		return p.GetRunning()
	case SessionStateTypeStopped:
		return p.GetStopped()
	}

	return nil
}
