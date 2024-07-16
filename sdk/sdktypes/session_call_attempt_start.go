package sdktypes

import (
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionCallAttemptStart struct {
	object[*SessionCallAttemptStartPB, SessionCallAttemptStartTraits]
}

var InvalidSessionCallAttemptStart SessionCallAttemptStart

type SessionCallAttemptStartPB = sessionv1.Call_Attempt_Start

type SessionCallAttemptStartTraits struct{}

func (SessionCallAttemptStartTraits) Validate(m *SessionCallAttemptStartPB) error { return nil }

func (SessionCallAttemptStartTraits) StrictValidate(m *SessionCallAttemptStartPB) error {
	return mandatory("started_at", m.StartedAt)
}

func SessionCallAttemptStartFromProto(m *SessionCallAttemptStartPB) (SessionCallAttemptStart, error) {
	return FromProto[SessionCallAttemptStart](m)
}

func StrictSessionCallAttemptStartFromProto(m *SessionCallAttemptStartPB) (SessionCallAttemptStart, error) {
	return Strict(SessionCallAttemptStartFromProto(m))
}

func (s SessionCallAttemptStart) Num() uint32 {
	return s.m.Num
}
