package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionCallAttemptResult struct {
	object[*SessionCallAttemptResultPB, SessionCallAttemptResultTraits]
}

var InvalidSessionCallAttemptResult SessionCallAttemptResult

type SessionCallAttemptResultPB = sessionv1.Call_Attempt_Result

type SessionCallAttemptResultTraits struct{ immutableObjectTrait }

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
	m := r.read()
	if m.Error == nil {
		return nil
	}

	return forceFromProto[ProgramError](m.Error).ToError()
}

func (r SessionCallAttemptResult) GetValue() Value {
	m := r.read()
	if m.Value == nil {
		return InvalidValue
	}

	return forceFromProto[Value](m.Value)
}

func (r SessionCallAttemptResult) ToPair() (Value, error) {
	return r.GetValue(), r.GetError()
}

func (r SessionCallAttemptResult) ToValueTuple() Value {
	err, v := r.GetError(), r.GetValue()

	if err == nil && !v.IsValid() {
		return InvalidValue
	}

	vs := []Value{Nothing, Nothing}

	if err != nil {
		vs[1] = WrapError(err).Value()
	} else {
		vs[0] = v
	}

	return kittehs.Must1(NewListValue(vs))
}
