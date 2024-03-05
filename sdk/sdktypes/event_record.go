package sdktypes

import (
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	eventv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
)

type EventRecord struct {
	object[*EventRecordPB, EventRecordTraits]
}

type EventRecordPB = eventv1.EventRecord

type EventRecordTraits struct{}

func (EventRecordTraits) Validate(m *EventRecordPB) error {
	return errors.Join(
		idField[EventID]("event_id", m.EventId),
		enumField[EventState]("state", m.State),
	)
}

func (t EventRecordTraits) StrictValidate(m *EventRecordPB) error {
	return errors.Join(
		mandatory("event_id", m.EventId),
		mandatory("state", m.State),
		mandatory("created_at", m.CreatedAt),
	)
}

func EventRecordFromProto(m *EventRecordPB) (EventRecord, error) { return FromProto[EventRecord](m) }
func StrictEventRecordFromProto(m *EventRecordPB) (EventRecord, error) {
	return Strict(EventRecordFromProto(m))
}

func (p EventRecord) EventID() (_ EventID) { return kittehs.Must1(ParseEventID(p.read().EventId)) }

func (p EventRecord) WithSeq(seq uint32) EventRecord {
	return EventRecord{p.forceUpdate(func(pb *EventRecordPB) { pb.Seq = seq })}
}

func (p EventRecord) WithCreatedAt(createdAt time.Time) EventRecord {
	return EventRecord{p.forceUpdate(func(pb *EventRecordPB) { pb.CreatedAt = timestamppb.New(createdAt) })}
}

func (p EventRecord) Seq() uint32       { return p.read().Seq }
func (p EventRecord) State() EventState { return forceEnumFromProto[EventState](p.read().State) }

func NewEventRecord(eventID EventID, state EventState) EventRecord {
	return kittehs.Must1(EventRecordFromProto(&EventRecordPB{
		EventId:   eventID.String(),
		State:     state.ToProto(),
		CreatedAt: timestamppb.New(time.Now()),
	}))
}
