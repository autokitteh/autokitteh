package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
)

func newRunStatusPB(s Object) *RunStatusPB {
	switch pb := s.toMessage().(type) {
	case *runtimesv1.RunStatus_Idle:
		return &RunStatusPB{States: &runtimesv1.RunStatus_Idle_{}}
	case *runtimesv1.RunStatus_Running:
		return &RunStatusPB{States: &runtimesv1.RunStatus_Running_{}}
	case *runtimesv1.RunStatus_LoadWait:
		return &RunStatusPB{States: &runtimesv1.RunStatus_LoadWait_{LoadWait: pb}}
	case *runtimesv1.CallWait:
		return &RunStatusPB{States: &runtimesv1.RunStatus_CallWait{CallWait: pb}}
	case *runtimesv1.RunStatus_Completed:
		return &RunStatusPB{States: &runtimesv1.RunStatus_Completed_{Completed: pb}}
	case *runtimesv1.RunStatus_Error:
		return &RunStatusPB{States: &runtimesv1.RunStatus_Error_{Error: pb}}
	default:
		sdklogger.Panic("unrecognized message type")
		return nil
	}
}

func getRunStatus(pb *RunStatusPB) (Object, error) {
	switch pb := pb.States.(type) {
	case *runtimesv1.RunStatus_Idle_:
		return idleRunState, nil
	case *runtimesv1.RunStatus_Running_:
		return runningRunState, nil
	case *runtimesv1.RunStatus_LoadWait_:
		return LoadWaitRunStateFromProto(pb.LoadWait)
	case *runtimesv1.RunStatus_Completed_:
		return CompletedRunStateFromProto(pb.Completed)
	case *runtimesv1.RunStatus_CallWait:
		return CallWaitRunStateFromProto(pb.CallWait)
	case *runtimesv1.RunStatus_Error_:
		return ErrorRunStateFromProto(pb.Error)
	default:
		return nil, fmt.Errorf("unrecognized type: %v, %w", pb, sdkerrors.ErrInvalidArgument)
	}
}

type (
	IdleRunState      = *object[*runtimesv1.RunStatus_Idle]
	RunningRunState   = *object[*runtimesv1.RunStatus_Running]
	LoadWaitRunState  = *object[*runtimesv1.RunStatus_LoadWait]
	CompletedRunState = *object[*runtimesv1.RunStatus_Completed]
	ErrorRunState     = *object[*runtimesv1.RunStatus_Error]

	// This is a bit special as it is reused in RunRequest.
	CallWaitRunState = *object[*runtimesv1.CallWait]
)

var (
	idleRunState     = NewRunStatus(makeMustFromProto[*runtimesv1.RunStatus_Idle](nil)(&runtimesv1.RunStatus_Idle{}))
	NewIdleRunStatus = func() RunStatus { return idleRunState }

	runningRunState     = NewRunStatus(makeMustFromProto[*runtimesv1.RunStatus_Running](nil)(&runtimesv1.RunStatus_Running{}))
	NewRunningRunStatus = func() RunStatus { return runningRunState }

	LoadWaitRunStateFromProto = makeFromProto[*runtimesv1.RunStatus_LoadWait](nil)
	NewLoadWaitRunStatus      = func(path string) RunStatus {
		return NewRunStatus(kittehs.Must1(LoadWaitRunStateFromProto(&runtimesv1.RunStatus_LoadWait{Path: path})))
	}

	CompletedRunStateFromProto = makeFromProto(validateCompletedRunState)
	NewCompletedRunStatus      = func(m map[string]Value) RunStatus {
		return NewRunStatus(kittehs.Must1(CompletedRunStateFromProto(&runtimesv1.RunStatus_Completed{
			Values: StringValueMapToProto(m),
		})))
	}

	ErrorRunStateFromProto = makeFromProto[*runtimesv1.RunStatus_Error](nil)
	NewErrorRunStatus      = func(errs []ProgramError) RunStatus {
		return NewRunStatus(kittehs.Must1(ErrorRunStateFromProto(&runtimesv1.RunStatus_Error{
			Errors: kittehs.Transform(errs, func(e ProgramError) *ProgramErrorPB { return e.ToProto() }),
		})))
	}

	CallWaitRunStateFromProto = makeFromProto(validateCallWaitRunState)
	NewCallWaitRunStatus      = func(call Value, args []Value, kwargs map[string]Value) RunStatus {
		return NewRunStatus(kittehs.Must1(CallWaitRunStateFromProto(&runtimesv1.CallWait{
			Call:   call.ToProto(),
			Args:   kittehs.Transform(args, func(v Value) *ValuePB { return v.ToProto() }),
			Kwargs: StringValueMapToProto(kwargs),
		})))
	}
)

func GetLoadWaitRunStatePath(l LoadWaitRunState) string { return l.pb.Path }

func validateCompletedRunState(pb *runtimesv1.RunStatus_Completed) error {
	_, err := kittehs.ValidateList(kittehs.MapValuesSortedByKeys(pb.Values), StrictValidateValuePB)
	return err
}

func validateCallWaitRunState(pb *runtimesv1.CallWait) error {
	if _, err := kittehs.ValidateList(kittehs.MapValuesSortedByKeys(pb.Kwargs), StrictValidateValuePB); err != nil {
		return fmt.Errorf("kwargs: %w", err)
	}

	if _, err := kittehs.ValidateList(pb.Args, StrictValidateValuePB); err != nil {
		return fmt.Errorf("args: %w", err)
	}

	if err := ValidateValuePB(pb.Call); err != nil {
		return fmt.Errorf("call: %w", err)
	}

	return nil
}
