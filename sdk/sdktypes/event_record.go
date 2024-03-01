package sdktypes

import (
	"fmt"
	"strings"
	"time"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type EventRecordPB = eventsv1.EventRecord

type EventRecord = *object[*EventRecordPB]

type EventState eventsv1.EventState

// Event Record State Enum
const (
	EventStateUnspecified = EventState(eventsv1.EventState_EVENT_STATE_UNSPECIFIED)
	EventStateSaved       = EventState(eventsv1.EventState_EVENT_STATE_SAVED)
	EventStateProcessing  = EventState(eventsv1.EventState_EVENT_STATE_PROCESSING)
	EventStateCompleted   = EventState(eventsv1.EventState_EVENT_STATE_COMPLETED)
	EventStateFailed      = EventState(eventsv1.EventState_EVENT_STATE_FAILED)
)

func (s EventState) String() string {
	return strings.TrimPrefix(eventsv1.EventState_name[int32(s)], "STATE_")
}

func (s EventState) ToProto() eventsv1.EventState {
	return eventsv1.EventState(s)
}

// Event Record

var (
	EventRecordFromProto       = makeFromProto(validateEventRecord)
	StrictEventRecordFromProto = makeFromProto(strictValidateEventRecord)
	ToStrictEventRecord        = makeWithValidator(strictValidateEventRecord)
)

func strictValidateEventRecord(pb *eventsv1.EventRecord) error {
	if err := ensureNotEmpty(pb.EventId); err != nil {
		return err
	}

	return validateEventRecord(pb)
}

func validateEventRecord(pb *eventsv1.EventRecord) error {
	if _, err := ParseEventID(pb.EventId); err != nil {
		return err
	}

	if _, ok := eventsv1.EventState_name[int32(pb.State)]; !ok {
		return fmt.Errorf("%w: invalid event state", sdkerrors.ErrInvalidArgument)
	}

	return nil
}

func GetEventRecordState(er EventRecord) EventState {
	return EventState(er.pb.State)
}

func GetEventRecordEventID(er EventRecord) EventID {
	if er == nil {
		return nil
	}

	return kittehs.Must1(ParseEventID(er.pb.EventId))
}

func GetEventRecordSeq(er EventRecord) uint32 {
	if er == nil {
		return 0
	}

	return er.pb.Seq
}

func ParseEventRecordState(raw string) EventState {
	if raw == "" {
		return EventStateUnspecified
	}
	upper := strings.ToUpper(raw)
	if !strings.HasPrefix(upper, "EVENT_STATE_") {
		upper = "EVENT_STATE_" + upper
	}

	state, ok := eventsv1.EventState_value[upper]
	if !ok {
		return EventStateUnspecified
	}

	return EventState(state)
}

var PossibleEventRecordStates = kittehs.Transform(kittehs.MapValuesSortedByKeys(eventsv1.EventState_name), func(name string) string {
	return strings.TrimPrefix(name, "EVENT_STATE_")
})

func GetEventRecordCreatedAt(er EventRecord) time.Time {
	if er == nil {
		return time.Time{}
	}

	return er.pb.CreatedAt.AsTime()
}
