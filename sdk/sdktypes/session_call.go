package sdktypes

import (
	"errors"

	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionCall struct {
	object[*SessionCallPB, SessionCallTraits]
}

func init() { registerObject[SessionCall]() }

var InvalidSessionCall SessionCall

type SessionCallPB = sessionv1.Call

type SessionCallTraits struct{ immutableObjectTrait }

func (SessionCallTraits) Validate(m *SessionCallPB) error {
	return errors.Join(
		objectField[SessionCallSpec]("spec", m.Spec),
		objectsSliceField[SessionCallAttempt]("attempt", m.Attempts),
	)
}

func (SessionCallTraits) StrictValidate(m *SessionCallPB) error {
	return mandatory("spec", m.Spec)
}

func SessionCallFromProto(m *SessionCallPB) (SessionCall, error) { return FromProto[SessionCall](m) }
func StrictSessionCallFromProto(m *SessionCallPB) (SessionCall, error) {
	return Strict(SessionCallFromProto(m))
}
