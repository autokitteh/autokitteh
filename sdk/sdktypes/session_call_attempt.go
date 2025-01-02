package sdktypes

import (
	"errors"

	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionCallAttempt struct {
	object[*SessionCallAttemptPB, SessionCallAttemptTraits]
}

var InvalidSessionCallAttempt SessionCallAttempt

type SessionCallAttemptPB = sessionv1.Call_Attempt

type SessionCallAttemptTraits struct{ immutableObjectTrait }

func (SessionCallAttemptTraits) Validate(m *SessionCallAttemptPB) error {
	return errors.Join(
		objectField[SessionCallAttemptStart]("start", m.Start),
		objectField[SessionCallAttemptComplete]("complete", m.Complete),
	)
}

func (SessionCallAttemptTraits) StrictValidate(m *SessionCallAttemptPB) error {
	return mandatory("start", m.Start)
}

func SessionCallAttemptFromProto(m *SessionCallAttemptPB) (SessionCallAttempt, error) {
	return FromProto[SessionCallAttempt](m)
}

func StrictSessionCallAttemptFromProto(m *SessionCallAttemptPB) (SessionCallAttempt, error) {
	return Strict(SessionCallAttemptFromProto(m))
}
