package sdktypes

import (
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionCallAttemptComplete struct {
	object[*SessionCallAttemptCompletePB, SessionCallAttemptCompleteTraits]
}

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

func NewSessionCallAttemptComplete(last bool, interval time.Duration, result SessionCallAttemptResult) SessionCallAttemptComplete {
	return forceFromProto[SessionCallAttemptComplete](&SessionCallAttemptCompletePB{
		IsLast:        last,
		RetryInterval: durationpb.New(interval),
		CompletedAt:   timestamppb.Now(),
		Result:        result.ToProto(),
	})
}
