package sdktypes

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
)

type (
	SessionStatePB = sessionsv1.SessionState // TODO: WrappedState?
	SessionState   = *object[*SessionStatePB]
)

var (
	SessionStateFromProto       = makeFromProto(validateSessionState)
	StrictSessionStateFromProto = makeFromProto(strictValidateSessionState)
	ToStrictSessionState        = makeWithValidator(strictValidateSessionState)
)

func strictValidateSessionState(pb *SessionStatePB) error {
	if pb.States == nil {
		return fmt.Errorf("no state")
	}

	return validateSessionState(pb)
}

func validateSessionState(pb *SessionStatePB) error {
	if _, err := unwrapSessionState(pb); err != nil {
		return err
	}
	return nil
}

func unwrapSessionState(pb *SessionStatePB) (Object, error) {
	switch pb := pb.States.(type) {
	case *sessionsv1.SessionState_Created_:
		return NewCreatedSessionState(), nil
	case *sessionsv1.SessionState_Running_:
		return RunningSessionStateFromProto(pb.Running)
	case *sessionsv1.SessionState_Completed_:
		return CompletedSessionStateFromProto(pb.Completed)
	case *sessionsv1.SessionState_Error_:
		return ErrorSessionStateFromProto(pb.Error)
	default:
		return nil, sdkerrors.ErrInvalidArgument
	}
}

func UnwrapSessionState(s SessionState) Object {
	return kittehs.Must1(unwrapSessionState(s.pb))
}

func WrapSessionState(s Object) SessionState {
	switch pb := s.toMessage().(type) {
	case *sessionsv1.SessionState_Created:
		return kittehs.Must1(SessionStateFromProto(&SessionStatePB{States: &sessionsv1.SessionState_Created_{Created: pb}}))
	case *sessionsv1.SessionState_Running:
		return kittehs.Must1(SessionStateFromProto(&SessionStatePB{States: &sessionsv1.SessionState_Running_{Running: pb}}))
	case *sessionsv1.SessionState_Completed:
		return kittehs.Must1(SessionStateFromProto(&SessionStatePB{States: &sessionsv1.SessionState_Completed_{Completed: pb}}))
	case *sessionsv1.SessionState_Error:
		return kittehs.Must1(SessionStateFromProto(&SessionStatePB{States: &sessionsv1.SessionState_Error_{Error: pb}}))
	default:
		sdklogger.DPanic("unhandled type")
		return nil
	}
}

func GetSessionStateType(s SessionState) SessionStateType {
	if s == nil {
		return UnspecifiedSessionStateType
	}

	switch ToProto(s).States.(type) {
	case *sessionsv1.SessionState_Created_:
		return CreatedSessionStateType
	case *sessionsv1.SessionState_Running_:
		return RunningSessionStateType
	case *sessionsv1.SessionState_Completed_:
		return CompletedSessionStateType
	case *sessionsv1.SessionState_Error_:
		return ErrorSessionStateType
	default:
		sdklogger.DPanic("unhandled type")
		return UnspecifiedSessionStateType
	}
}

type (
	CreatedSessionState   = *object[*sessionsv1.SessionState_Created]
	RunningSessionState   = *object[*sessionsv1.SessionState_Running]
	ErrorSessionState     = *object[*sessionsv1.SessionState_Error]
	CompletedSessionState = *object[*sessionsv1.SessionState_Completed]
)

var (
	createdSessionState    = makeMustFromProto[*sessionsv1.SessionState_Created](nil)(&sessionsv1.SessionState_Created{})
	NewCreatedSessionState = func() CreatedSessionState { return createdSessionState }

	RunningSessionStateFromProto = makeFromProto(validateRunningSessionState)
	NewRunningSessionState       = func(runID RunID, callv Value) RunningSessionState {
		return kittehs.Must1(RunningSessionStateFromProto(&sessionsv1.SessionState_Running{
			RunId: runID.String(),
			Call:  ToProto(callv),
		}))
	}

	ErrorSessionStateFromProto = makeFromProto(validateErrorSessionState)
	NewErrorSessionState       = func(err error, prints []string) ErrorSessionState {
		return kittehs.Must1(
			ErrorSessionStateFromProto(
				&sessionsv1.SessionState_Error{
					Error:  ProgramErrorFromError(err).ToProto(),
					Prints: prints,
				},
			),
		)
	}

	CompletedSessionStateFromProto = makeFromProto(validateCompletedSessionState)
	NewCompletedSessionState       = func(prints []string, exports map[string]Value, ret Value) (CompletedSessionState, error) {
		return CompletedSessionStateFromProto(&sessionsv1.SessionState_Completed{
			Prints:      prints,
			Exports:     kittehs.TransformMapValues(exports, ToProto),
			ReturnValue: ret.ToProto(),
		})
	}
)

func validateRunningSessionState(v *sessionsv1.SessionState_Running) error {
	if _, err := ParseRunID(v.RunId); err != nil {
		return fmt.Errorf("run_id: %w", err)
	}

	if v.Call != nil {
		callv, err := ValueFromProto(v.Call)
		if err != nil {
			return fmt.Errorf("call: %w", err)
		}

		if !IsFunctionValue(callv) {
			return fmt.Errorf("call value is not a function")
		}
	}

	return nil
}

func validateErrorSessionState(v *sessionsv1.SessionState_Error) error {
	if err := validateProgramError(v.Error); err != nil {
		return fmt.Errorf("error: %w", err)
	}

	return nil
}

func validateCompletedSessionState(v *sessionsv1.SessionState_Completed) error {
	if err := kittehs.ValidateMap(v.Exports, func(k string, v *ValuePB) error {
		if _, err := ParseSymbol(k); err != nil {
			return fmt.Errorf("key %q: %w", k, err)
		}

		if err := ValidateValuePB(v); err != nil {
			return fmt.Errorf("value: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("exports: %w", err)
	}

	if err := ValidateValuePB(v.ReturnValue); err != nil {
		return fmt.Errorf("return_value: %w", err)
	}

	return nil
}

func SessionStateWithTimestamp(s SessionState, t time.Time) SessionState {
	return kittehs.Must1(s.Update(func(pb *SessionStatePB) {
		pb.T = timestamppb.New(t)
	}))
}

func GetSessionStateRunID(s SessionState) RunID {
	return kittehs.Must1(ParseRunID(s.pb.States.(*sessionsv1.SessionState_Running_).Running.RunId))
}

func GetSessionStateCallValue(s SessionState) Value {
	return kittehs.Must1(ValueFromProto(s.pb.States.(*sessionsv1.SessionState_Running_).Running.Call))
}

func GetSessionHistoryStatePrints(s SessionState) []string {
	switch pb := s.pb.States.(type) {
	case *sessionsv1.SessionState_Completed_:
		return pb.Completed.Prints
	case *sessionsv1.SessionState_Error_:
		return pb.Error.Prints
	default:
		return nil
	}
}
