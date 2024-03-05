package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
)

type eventStateTraits struct{}

var _ enumTraits = eventStateTraits{}

func (eventStateTraits) Prefix() string           { return "EVENT_RECORD_STATE_" }
func (eventStateTraits) Names() map[int32]string  { return eventsv1.EventState_name }
func (eventStateTraits) Values() map[string]int32 { return eventsv1.EventState_value }

type EventState struct {
	enum[eventStateTraits, eventsv1.EventState]
}

func eventStateFromProto(e eventsv1.EventState) EventState {
	return kittehs.Must1(EventStateFromProto(e))
}

var (
	PossibleEventStatesNames = AllEnumNames[eventStateTraits]()

	EventStateUnspecified = eventStateFromProto(eventsv1.EventState_EVENT_STATE_UNSPECIFIED)
	EventStateSaved       = eventStateFromProto(eventsv1.EventState_EVENT_STATE_SAVED)
	EventStateProcessing  = eventStateFromProto(eventsv1.EventState_EVENT_STATE_PROCESSING)
	EventStateCompleted   = eventStateFromProto(eventsv1.EventState_EVENT_STATE_COMPLETED)
	EventStateFailed      = eventStateFromProto(eventsv1.EventState_EVENT_STATE_FAILED)
)

func EventStateFromProto(e eventsv1.EventState) (EventState, error) {
	return EnumFromProto[EventState](e)
}

func ParseEventState(raw string) (EventState, error) {
	return ParseEnum[EventState](raw)
}
