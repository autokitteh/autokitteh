package sdktypes

import (
	"errors"
	"time"

	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
)

type SessionState struct {
	object[*SessionStatePB, SessionStateTraits]
}

type SessionStatePB = sessionv1.SessionState

type SessionStateTraits struct{}

func (SessionStateTraits) Validate(m *SessionStatePB) error {
	return errors.Join(
		objectField[SessionStateCompleted]("completed", m.Completed),
		objectField[SessionStateCreated]("created", m.Created),
		objectField[SessionStateError]("error", m.Error),
		objectField[SessionStateRunning]("running", m.Running),
	)
}

func (SessionStateTraits) StrictValidate(m *SessionStatePB) error {
	return oneOfMessage(m)
}

func SessionStateFromProto(m *SessionStatePB) (SessionState, error) {
	return FromProto[SessionState](m)
}

func StrictSessionStateFromProto(m *SessionStatePB) (SessionState, error) {
	return Strict(SessionStateFromProto(m))
}

func NewSessionState(t time.Time, concrete concreteSessionState) SessionState {
	var pb SessionStatePB

	switch concrete := concrete.(type) {
	case *SessionStateCompleted:
		pb.Completed = concrete.ToProto()
	case *SessionStateCreated:
		pb.Created = concrete.ToProto()
	case *SessionStateError:
		pb.Error = concrete.ToProto()
	default:
		sdklogger.Panic("invalid session concrete state")
	}

	return forceFromProto[SessionState](&pb)
}

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
	}

	return nil
}
