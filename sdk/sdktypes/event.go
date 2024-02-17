package sdktypes

import (
	"fmt"
	"time"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
)

type EventPB = eventsv1.Event

type Event = *object[*EventPB]

var (
	EventFromProto       = makeFromProto(validateEvent)
	StrictEventFromProto = makeFromProto(strictValidateEvent)
	ToStrictEvent        = makeWithValidator(strictValidateEvent)
)

func strictValidateEvent(pb *eventsv1.Event) error {
	if err := ensureNotEmpty(pb.EventId, pb.EventType, pb.IntegrationId, pb.IntegrationToken); err != nil {
		return err
	}

	return validateEvent(pb)
}

func validateEvent(pb *eventsv1.Event) error {
	if _, err := ParseEventID(pb.EventId); err != nil {
		return fmt.Errorf("event id: %w", err)
	}

	if _, err := ParseIntegrationID(pb.IntegrationId); err != nil {
		return fmt.Errorf("integration id: %w", err)
	}

	if _, err := kittehs.TransformMapValuesError(pb.Data, ValueFromProto); err != nil {
		return fmt.Errorf("data: %w", err)
	}

	return nil
}

func GetEventID(e Event) EventID {
	if e == nil {
		return nil
	}
	return kittehs.Must1(ParseEventID(e.pb.EventId))
}

func GetEventIntegrationID(e Event) IntegrationID {
	if e == nil {
		return nil
	}
	return kittehs.Must1(ParseIntegrationID(e.pb.IntegrationId))
}

func GetEventIntegrationToken(e Event) string {
	if e == nil {
		return ""
	}
	return e.pb.IntegrationToken
}

func GetEventMemo(e Event) map[string]string {
	if e == nil {
		return nil
	}
	return e.pb.Memo
}

func GetEventType(e Event) string {
	if e == nil {
		return ""
	}
	return e.pb.EventType
}

func GetEventData(e Event) map[string]Value {
	if e == nil {
		return nil
	}
	return kittehs.TransformMapValues(e.pb.Data, MustValueFromProto)
}

func GetOriginalEventID(e Event) string {
	if e == nil {
		return ""
	}
	return e.pb.OriginalEventId
}

func GetEventCreatedAt(e Event) time.Time {
	if e == nil {
		return time.Time{}
	}

	return e.pb.CreatedAt.AsTime()
}

func GetEventSequenceNumber(e Event) uint64 {
	if e == nil {
		return 0
	}

	return e.pb.Seq
}

func EventToValues(e Event) map[string]Value {
	return map[string]Value{
		"event_type":        NewStringValue(GetEventType(e)),
		"event_id":          NewStringValue(GetEventID(e).String()),
		"original_event_id": NewStringValue(GetOriginalEventID(e)),
		"integration_id":    NewStringValue(GetEventIntegrationID(e).String()),
		"data": NewStructValue(
			NewStringValue("event_data"),
			kittehs.FilterMapKeys(GetEventData(e), func(k string) bool {
				_, err := StrictParseSymbol(k)
				return err == nil
			}),
		),
	}
}
