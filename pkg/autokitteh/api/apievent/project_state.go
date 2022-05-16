package apievent

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"

	pbevent "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/event"
)

type ProjectEventState struct{ pb *pbevent.ProjectEventState }

func (s *ProjectEventState) PB() *pbevent.ProjectEventState {
	return proto.Clone(s.pb).(*pbevent.ProjectEventState)
}
func (s *ProjectEventState) Clone() *ProjectEventState { return &ProjectEventState{pb: s.PB()} }

func ProjectEventStateFromProto(pb *pbevent.ProjectEventState) (*ProjectEventState, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&ProjectEventState{pb: pb}).Clone(), nil
}

func MustProjectEventStateFromProto(pb *pbevent.ProjectEventState) *ProjectEventState {
	s, err := ProjectEventStateFromProto(pb)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *ProjectEventState) Name() string { return s.Get().Name() }

func (s *ProjectEventState) IsError() bool {
	if s == nil || s.pb == nil {
		return false
	}

	_, ok := s.pb.Type.(*pbevent.ProjectEventState_Error)

	return ok
}

func (s *ProjectEventState) Get() projectEventState {
	if s == nil || s.pb == nil {
		return nil
	}

	switch pb := s.pb.Type.(type) {
	case *pbevent.ProjectEventState_Error:
		return &ErrorProjectEventState{pb: pb.Error}
	case *pbevent.ProjectEventState_Ignored:
		return &IgnoredProjectEventState{pb: pb.Ignored}
	case *pbevent.ProjectEventState_Pending:
		return &PendingProjectEventState{pb: pb.Pending}
	case *pbevent.ProjectEventState_Processing:
		return &ProcessingProjectEventState{pb: pb.Processing}
	case *pbevent.ProjectEventState_Waiting:
		return &WaitingProjectEventState{pb: pb.Waiting}
	case *pbevent.ProjectEventState_Processed:
		return &ProcessedProjectEventState{pb: pb.Processed}
	default:
		panic(fmt.Errorf("unrecognized event state: %v", reflect.TypeOf(s.pb.Type)))
	}
}

func NewProjectEventState(state projectEventState) *ProjectEventState {
	var pb pbevent.ProjectEventState

	switch s := state.(type) {
	case *ErrorProjectEventState:
		pb.Type = &pbevent.ProjectEventState_Error{Error: s.pb}
	case *IgnoredProjectEventState:
		pb.Type = &pbevent.ProjectEventState_Ignored{Ignored: s.pb}
	case *PendingProjectEventState:
		pb.Type = &pbevent.ProjectEventState_Pending{Pending: s.pb}
	case *ProcessingProjectEventState:
		pb.Type = &pbevent.ProjectEventState_Processing{Processing: s.pb}
	case *ProcessedProjectEventState:
		pb.Type = &pbevent.ProjectEventState_Processed{Processed: s.pb}
	case *WaitingProjectEventState:
		pb.Type = &pbevent.ProjectEventState_Waiting{Waiting: s.pb}
	default:
		panic(fmt.Errorf("unrecognized event state: %v", reflect.TypeOf(s)))
	}

	return MustProjectEventStateFromProto(&pb)
}
