package apilang

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"

	pblang "github.com/autokitteh/autokitteh/api/gen/stubs/go/lang"
)

type RunState struct{ pb *pblang.RunState }

func (r *RunState) PB() *pblang.RunState {
	if r == nil || r.pb == nil {
		return nil
	}

	return proto.Clone(r.pb).(*pblang.RunState)
}

func (r *RunState) Clone() *RunState { return &RunState{pb: r.PB()} }

func (r *RunState) IsState() bool {
	if r == nil || r.pb == nil {
		return false
	}

	return r.Get().IsRunState()
}

func (r *RunState) IsRunning() bool {
	if r == nil || r.pb == nil {
		return false
	}

	_, ok := r.Get().(*RunningRunState)
	return ok
}

func (r *RunState) IsFinal() bool {
	if r == nil || r.pb == nil {
		return false
	}

	return r.Get().IsFinal()
}

func (r *RunState) IsDiscardable() bool {
	if r == nil || r.pb == nil {
		return false
	}

	return r.Get().IsDiscardable()
}

func (r *RunState) Name() string {
	if r == nil || r.pb == nil {
		return "none"
	}

	return r.Get().Name()
}

func RunStateFromProto(pb *pblang.RunState) (*RunState, error) {
	if pb == nil {
		return nil, nil
	}

	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?

	return (&RunState{pb: pb}).Clone(), nil
}

func MustRunStateFromProto(pb *pblang.RunState) *RunState {
	s, err := RunStateFromProto(pb)
	if err != nil {
		panic(err)
	}
	return s
}

func (r *RunState) Get() runState {
	if r == nil || r.pb == nil {
		return nil
	}

	switch pb := r.pb.Type.(type) {
	case *pblang.RunState_Running:
		return &RunningRunState{pb: pb.Running}
	case *pblang.RunState_Canceled:
		return &CanceledRunState{pb: pb.Canceled}
	case *pblang.RunState_Call:
		return &CallWaitRunState{pb: pb.Call}
	case *pblang.RunState_Load:
		return &LoadWaitRunState{pb: pb.Load}
	case *pblang.RunState_Error:
		return &ErrorRunState{pb: pb.Error}
	case *pblang.RunState_Completed:
		return &CompletedRunState{pb: pb.Completed}
	case *pblang.RunState_Loadret:
		return &LoadReturnedRunUpdate{pb: pb.Loadret}
	case *pblang.RunState_Print:
		return &PrintRunUpdate{pb: pb.Print}
	case *pblang.RunState_Callret:
		return &CallReturnedRunUpdate{pb: pb.Callret}
	case *pblang.RunState_ClientError:
		return &ClientErrorRunState{pb: pb.ClientError}
	default:
		panic(fmt.Errorf("unrecognized run state: %v", reflect.TypeOf(r.pb.Type)))
	}
}

func NewRunState(state runState) (*RunState, error) {
	var pb pblang.RunState

	switch s := state.(type) {
	case *RunningRunState:
		pb.Type = &pblang.RunState_Running{Running: s.pb}
	case *CanceledRunState:
		pb.Type = &pblang.RunState_Canceled{Canceled: s.pb}
	case *CallWaitRunState:
		pb.Type = &pblang.RunState_Call{Call: s.pb}
	case *LoadWaitRunState:
		pb.Type = &pblang.RunState_Load{Load: s.pb}
	case *ErrorRunState:
		pb.Type = &pblang.RunState_Error{Error: s.pb}
	case *CompletedRunState:
		pb.Type = &pblang.RunState_Completed{Completed: s.pb}
	case *PrintRunUpdate:
		pb.Type = &pblang.RunState_Print{Print: s.pb}
	case *CallReturnedRunUpdate:
		pb.Type = &pblang.RunState_Callret{Callret: s.pb}
	case *LoadReturnedRunUpdate:
		pb.Type = &pblang.RunState_Loadret{Loadret: s.pb}
	case *ClientErrorRunState:
		pb.Type = &pblang.RunState_ClientError{ClientError: s.pb}
	default:
		panic(fmt.Errorf("unrecognized run state: %w", reflect.TypeOf(s)))
	}

	return RunStateFromProto(&pb)
}

func MustNewRunState(state runState) *RunState {
	s, err := NewRunState(state)
	if err != nil {
		panic(err)
	}
	return s
}
