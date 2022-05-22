package apievent

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"

	pbevent "github.com/autokitteh/autokitteh/api/gen/stubs/go/event"
)

type EventState struct{ pb *pbevent.EventState }

func (s *EventState) PB() *pbevent.EventState { return proto.Clone(s.pb).(*pbevent.EventState) }
func (s *EventState) Clone() *EventState      { return &EventState{pb: s.PB()} }

func EventStateFromProto(pb *pbevent.EventState) (*EventState, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&EventState{pb: pb}).Clone(), nil
}

func MustEventStateFromProto(pb *pbevent.EventState) *EventState {
	s, err := EventStateFromProto(pb)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *EventState) Name() string { return s.Get().Name() }

func (s *EventState) Get() eventState {
	if s == nil || s.pb == nil {
		return nil
	}

	switch pb := s.pb.Type.(type) {
	case *pbevent.EventState_Error:
		return &ErrorEventState{pb: pb.Error}
	case *pbevent.EventState_Ignored:
		return &IgnoredEventState{pb: pb.Ignored}
	case *pbevent.EventState_Pending:
		return &PendingEventState{pb: pb.Pending}
	case *pbevent.EventState_Processing:
		return &ProcessingEventState{pb: pb.Processing}
	case *pbevent.EventState_Processed:
		return &ProcessedEventState{pb: pb.Processed}
	default:
		panic(fmt.Errorf("unrecognized event state: %v", reflect.TypeOf(s.pb.Type)))
	}
}

func NewEventState(state eventState) *EventState {
	var pb pbevent.EventState

	switch s := state.(type) {
	case *ErrorEventState:
		pb.Type = &pbevent.EventState_Error{Error: s.pb}
	case *IgnoredEventState:
		pb.Type = &pbevent.EventState_Ignored{Ignored: s.pb}
	case *PendingEventState:
		pb.Type = &pbevent.EventState_Pending{Pending: s.pb}
	case *ProcessingEventState:
		pb.Type = &pbevent.EventState_Processing{Processing: s.pb}
	case *ProcessedEventState:
		pb.Type = &pbevent.EventState_Processed{Processed: s.pb}
	default:
		panic(fmt.Errorf("unrecognized event state: %w", reflect.TypeOf(s)))
	}

	return MustEventStateFromProto(&pb)
}
