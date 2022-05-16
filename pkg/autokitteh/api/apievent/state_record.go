package apievent

import (
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbevent "github.com/autokitteh/autokitteh/gen/proto/stubs/go/event"
)

type EventStateRecord struct{ pb *pbevent.EventStateRecord }

func (sr *EventStateRecord) PB() *pbevent.EventStateRecord {
	return proto.Clone(sr.pb).(*pbevent.EventStateRecord)
}

func (sr *EventStateRecord) Clone() *EventStateRecord { return &EventStateRecord{pb: sr.PB()} }

func (sr *EventStateRecord) T() time.Time       { return sr.pb.T.AsTime() }
func (sr *EventStateRecord) State() *EventState { return MustEventStateFromProto(sr.pb.State) }

func EventStateRecordFromProto(pb *pbevent.EventStateRecord) (*EventStateRecord, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&EventStateRecord{pb: pb}).Clone(), nil
}

func NewEventStateRecord(s *EventState, t time.Time) (*EventStateRecord, error) {
	return EventStateRecordFromProto(&pbevent.EventStateRecord{
		State: s.PB(),
		T:     timestamppb.New(t),
	})
}
