package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionStateCompleted struct {
	object[*SessionStateCompletedPB, SessionStateCompletedTraits]
}

func init() { registerObject[SessionStateCompleted]() }

var InvalidSessionStateCompleted SessionStateCompleted

type SessionStateCompletedPB = sessionv1.SessionState_Completed

type SessionStateCompletedTraits struct{ immutableObjectTrait }

func (SessionStateCompleted) isConcreteSessionState() {}

func (SessionStateCompletedTraits) Validate(m *SessionStateCompletedPB) error {
	return errors.Join(
		objectField[Value]("value", m.ReturnValue),
		valuesMapField("exports", m.Exports),
	)
}

func (SessionStateCompletedTraits) StrictValidate(m *SessionStateCompletedPB) error {
	return mandatory("return_value", m.ReturnValue)
}

func SessionStateCompletedFromProto(m *SessionStateCompletedPB) (SessionStateCompleted, error) {
	return FromProto[SessionStateCompleted](m)
}

func StrictSessionStateCompletedFromProto(m *SessionStateCompletedPB) (SessionStateCompleted, error) {
	return Strict(SessionStateCompletedFromProto(m))
}

func (s SessionState) GetCompleted() SessionStateCompleted {
	return forceFromProto[SessionStateCompleted](s.read().Completed)
}

func NewSessionStateCompleted(prints []string, exports map[string]Value, ret Value) SessionState {
	return forceFromProto[SessionState](&sessionv1.SessionState{
		Completed: &SessionStateCompletedPB{
			Prints:      prints,
			Exports:     kittehs.TransformMapValues(exports, ToProto),
			ReturnValue: ret.ToProto(),
		},
	})
}
