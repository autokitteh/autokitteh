package sdktypes

import (
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionCallAttemptComplete struct {
	object[*SessionCallAttemptCompletePB, SessionCallAttemptCompleteTraits]
}

func init() { registerObject[SessionCallAttemptComplete]() }

var InvalidSessionCallAttemptComplete SessionCallAttemptComplete

type SessionCallAttemptCompletePB = sessionv1.Call_Attempt_Complete

type SessionCallAttemptCompleteTraits struct{ immutableObjectTrait }

func (SessionCallAttemptCompleteTraits) Validate(m *SessionCallAttemptCompletePB) error {
	return objectField[SessionCallAttemptResult]("result", m.Result)
}

func (SessionCallAttemptCompleteTraits) StrictValidate(m *SessionCallAttemptCompletePB) error {
	return errors.Join(
		mandatory("completed_at", m.CompletedAt),
		mandatory("result", m.Result),
	)
}

func SessionCallAttemptCompleteFromProto(m *SessionCallAttemptCompletePB) (SessionCallAttemptComplete, error) {
	return FromProto[SessionCallAttemptComplete](m)
}

func StrictSessionCallAttemptCompleteFromProto(m *SessionCallAttemptCompletePB) (SessionCallAttemptComplete, error) {
	return Strict(SessionCallAttemptCompleteFromProto(m))
}

func (p SessionCallAttemptComplete) Result() SessionCallAttemptResult {
	return forceFromProto[SessionCallAttemptResult](p.read().Result)
}

func NewSessionLogCallAttemptComplete(complete SessionCallAttemptComplete) SessionCallAttemptComplete {
	return forceFromProto[SessionCallAttemptComplete](&SessionCallAttemptCompletePB{Result: ToProto(complete.Result())})
}

func NewSessionCallAttemptComplete(t time.Time, last bool, result SessionCallAttemptResult) SessionCallAttemptComplete {
	return forceFromProto[SessionCallAttemptComplete](&SessionCallAttemptCompletePB{
		IsLast:      last,
		CompletedAt: timestamppb.New(t),
		Result:      result.ToProto(),
	})
}
