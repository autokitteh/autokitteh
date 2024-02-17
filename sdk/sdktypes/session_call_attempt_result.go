package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type (
	SessionCallAttemptResultPB = sessionsv1.Call_Attempt_Result
	SessionCallAttemptResult   = *object[*SessionCallAttemptResultPB]
)

var (
	SessionCallAttemptResultFromProto       = makeFromProto(validateSessionCallAttemptResult)
	StrictSessionCallAttemptResultFromProto = makeFromProto(strictValidateSessionCallAttemptResult)
	ToStrictSessionCallAttemptResult        = makeWithValidator(strictValidateSessionCallAttemptResult)
)

func strictValidateSessionCallAttemptResult(pb *SessionCallAttemptResultPB) error {
	return validateSessionCallAttemptResult(pb)
}

func validateSessionCallAttemptResult(pb *SessionCallAttemptResultPB) error {
	switch v := pb.GetResult().(type) {
	case *sessionsv1.Call_Attempt_Result_Value:
		return ValidateValuePB(v.Value)
	case *sessionsv1.Call_Attempt_Result_Error:
		return nil
	default:
		return fmt.Errorf("unknown session call result type: %T", v)
	}
}

func NewSessionCallAttemptResult(v Value, err error) SessionCallAttemptResult {
	if err != nil {
		return NewSessionCallAttemptErrorResult(err)
	}

	if v != nil {
		return NewSessionCallAttemptValueResult(v)
	}

	return nil
}

func NewSessionCallAttemptValueResult(v Value) SessionCallAttemptResult {
	return kittehs.Must1(SessionCallAttemptResultFromProto(&SessionCallAttemptResultPB{
		Result: &sessionsv1.Call_Attempt_Result_Value{
			Value: v.ToProto(),
		},
	}))
}

func NewSessionCallAttemptErrorResult(err error) SessionCallAttemptResult {
	return kittehs.Must1(SessionCallAttemptResultFromProto(&SessionCallAttemptResultPB{
		Result: &sessionsv1.Call_Attempt_Result_Error{
			Error: ProgramErrorFromError(err).ToProto(),
		},
	}))
}

func SessionCallResultAsPair(r SessionCallAttemptResult) (Value, error) {
	if r == nil {
		return nil, nil
	}

	switch v := r.pb.GetResult().(type) {
	case *sessionsv1.Call_Attempt_Result_Value:
		return kittehs.Must1(ValueFromProto(v.Value)), nil
	case *sessionsv1.Call_Attempt_Result_Error:
		return nil, ProgramErrorToError(kittehs.Must1(ProgramErrorFromProto(v.Error)))
	default:
		return nil, fmt.Errorf("unknown session call result type: %T", v)
	}
}

func GetSessionCallResultError(r SessionCallAttemptResult) error {
	if r == nil {
		return nil
	}

	switch v := r.pb.GetResult().(type) {
	case *sessionsv1.Call_Attempt_Result_Error:
		return ProgramErrorToError(kittehs.Must1(ProgramErrorFromProto(v.Error)))
	default:
		return nil
	}
}

func GetSessionCallResultValue(r SessionCallAttemptResult) Value {
	if r == nil {
		return nil
	}

	switch v := r.pb.GetResult().(type) {
	case *sessionsv1.Call_Attempt_Result_Value:
		return kittehs.Must1(ValueFromProto(v.Value))
	default:
		return nil
	}
}
