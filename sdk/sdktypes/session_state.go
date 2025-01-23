package sdktypes

import (
	"errors"
	"fmt"
	"time"

	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
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
	case *SessionStateStopped:
		pb.Stopped = concrete.ToProto()
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

var sessionStateStructCtor = NewSymbolValue(NewSymbol("session_state"))

func (p SessionState) ToValue() (Value, error) {
	fields := make(map[string]Value)

	switch s := p.Concrete().(type) {
	case SessionStateCompleted:
		fields["value"] = s.ReturnValue()
	case SessionStateCreated:
	case SessionStateError:
		fields["value"] = s.GetProgramError().Value()
	case SessionStateRunning:
	case SessionStateStopped:
		fields["reason"] = NewStringValue(s.Reason())
	default:
		return InvalidValue, fmt.Errorf("%w: unknown state type", sdkerrors.ErrUnretryableUnknown)
	}

	fields["type"] = NewStringValue(p.Type().String())

	return NewStructValue(sessionStateStructCtor, fields)
}
