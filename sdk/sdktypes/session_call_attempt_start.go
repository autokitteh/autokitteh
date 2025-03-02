package sdktypes

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionCallAttemptStart struct {
	object[*SessionCallAttemptStartPB, SessionCallAttemptStartTraits]
}

func init() { registerObject[SessionCallAttemptStart]() }

var InvalidSessionCallAttemptStart SessionCallAttemptStart

type SessionCallAttemptStartPB = sessionv1.Call_Attempt_Start

type SessionCallAttemptStartTraits struct{ immutableObjectTrait }

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

func NewSessionCallAttemptStart(t time.Time, n uint32) SessionCallAttemptStart {
	return kittehs.Must1(SessionCallAttemptStartFromProto(&SessionCallAttemptStartPB{
		StartedAt: timestamppb.New(t),
		Num:       n,
	}))
}
