package sdktypes

import (
	"errors"

	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionCallAttemptResult struct {
	object[*SessionCallAttemptResultPB, SessionCallAttemptResultTraits]
}

var InvalidSessionCallAttemptResult SessionCallAttemptResult

type SessionCallAttemptResultPB = sessionv1.Call_Attempt_Result

type SessionCallAttemptResultTraits struct{}

func (SessionCallAttemptResultTraits) Validate(m *SessionCallAttemptResultPB) error {
	return errors.Join(
		objectField[ProgramError]("error", m.Error),
		objectField[Value]("value", m.Value),
	)
}

func (SessionCallAttemptResultTraits) StrictValidate(m *SessionCallAttemptResultPB) error {
	return nonzeroMessage(m)
}

func SessionCallAttemptResultFromProto(m *SessionCallAttemptResultPB) (SessionCallAttemptResult, error) {
	return FromProto[SessionCallAttemptResult](m)
}

func StrictSessionCallAttemptResultFromProto(m *SessionCallAttemptResultPB) (SessionCallAttemptResult, error) {
	return Strict(SessionCallAttemptResultFromProto(m))
}

func NewSessionCallAttemptResult(v Value, err error) SessionCallAttemptResult {
	var pb SessionCallAttemptResultPB

	if err != nil {
		pb.Error = WrapError(err).ToProto()
	} else {
		pb.Value = v.ToProto()
	}

	return forceFromProto[SessionCallAttemptResult](&pb)
}

func (r SessionCallAttemptResult) GetError() error {
	return forceFromProto[ProgramError](r.read().Error).ToError()
}

func (r SessionCallAttemptResult) GetValue() Value {
	return forceFromProto[Value](r.read().Value)
}

func (r SessionCallAttemptResult) ToPair() (Value, error) {
	return r.GetValue(), r.GetError()
}
