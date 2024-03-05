package sdktypes

import (
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	eventv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
)

type Event struct{ object[*EventPB, EventTraits] }

var InvalidEvent Event

type EventPB = eventv1.Event

type EventTraits struct{}

func (EventTraits) Validate(m *EventPB) error {
	return errors.Join(
		idField[EventID]("event_id", m.EventId),
		idField[IntegrationID]("integration_id", m.IntegrationId),
	)
}

func (EventTraits) StrictValidate(m *EventPB) error {
	return errors.Join(
		mandatory("event_id", m.EventId),
		mandatory("event_type", m.EventType),
		mandatory("integration_id", m.IntegrationId),
		mandatory("created_at", m.CreatedAt),
	)
}

func EventFromProto(m *EventPB) (Event, error)       { return FromProto[Event](m) }
func StrictEventFromProto(m *EventPB) (Event, error) { return Strict(EventFromProto(m)) }

func (p Event) ID() EventID { return kittehs.Must1(ParseEventID(p.read().EventId)) }

func (e Event) WithNewID() Event {
	return Event{e.forceUpdate(func(m *EventPB) { m.EventId = NewEventID().String() })}
}

func (e Event) WithCreatedAt(t time.Time) Event {
	return Event{e.forceUpdate(func(m *EventPB) { m.CreatedAt = timestamppb.New(t) })}
}

func (e Event) WithMemo(memo map[string]string) Event {
	return Event{e.forceUpdate(func(m *EventPB) { m.Memo = memo })}
}

func (e Event) Memo() map[string]string { return e.read().Memo }

func (e Event) IntegrationID() IntegrationID {
	return kittehs.Must1(ParseIntegrationID(e.read().IntegrationId))
}

func (e Event) IntegrationToken() string { return e.read().IntegrationToken }

func (e Event) Type() string { return e.read().EventType }

func (e Event) ToValues() map[string]Value {
	pb := e.read()

	return map[string]Value{
		"event_type":        NewStringValue(pb.EventType),
		"event_id":          NewStringValue(pb.EventId),
		"original_event_id": NewStringValue(pb.OriginalEventId),
		"integration_id":    NewStringValue(pb.IntegrationId),
		"data": kittehs.Must1(NewStructValue(
			NewStringValue("event_data"),
			kittehs.TransformMapValues(
				kittehs.FilterMapKeys(pb.Data, IsValidSymbol),
				forceFromProto[Value],
			),
		)),
	}
}

func (e Event) OriginalEventID() string { return e.read().OriginalEventId }
func (e Event) CreatedAt() time.Time    { return e.read().CreatedAt.AsTime() }
func (e Event) Data() map[string]Value {
	return kittehs.TransformMapValues(e.read().Data, forceFromProto[Value])
}
func (e Event) Seq() uint64 { return e.read().Seq }
