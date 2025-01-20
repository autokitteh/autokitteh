package sdktypes

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/cel-go/cel"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	eventv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
)

type Event struct{ object[*EventPB, EventTraits] }

func init() { registerObject[Event]() }

var InvalidEvent Event

type EventPB = eventv1.Event

type EventTraits struct{ immutableObjectTrait }

func (EventTraits) Validate(m *EventPB) error {
	return errors.Join(
		idField[EventID]("event_id", m.EventId),
	)
}

func (EventTraits) StrictValidate(m *EventPB) error {
	return errors.Join(
		mandatory("event_id", m.EventId),
		mandatory("created_at", m.CreatedAt),
		mandatory("destination_id", m.DestinationId),
	)
}

func NewEvent[T concreteEventDestinationID](dstID T) Event {
	return kittehs.Must1(EventFromProto(&EventPB{})).
		WithNewID().
		WithDestinationID(NewEventDestinationID(dstID))
}

func EventFromProto(m *EventPB) (Event, error)       { return FromProto[Event](m) }
func StrictEventFromProto(m *EventPB) (Event, error) { return Strict(EventFromProto(m)) }

func (p Event) ID() EventID { return kittehs.Must1(ParseEventID(p.read().EventId)) }

func (e Event) WithID(id EventID) Event {
	return Event{e.forceUpdate(func(m *EventPB) { m.EventId = id.String() })}
}

func (e Event) WithNewID() Event { return e.WithID(NewEventID()) }

func (e Event) WithCreatedAt(t time.Time) Event {
	return Event{e.forceUpdate(func(m *EventPB) { m.CreatedAt = timestamppb.New(t) })}
}

func (e Event) WithMemo(memo map[string]string) Event {
	return Event{e.forceUpdate(func(m *EventPB) { m.Memo = memo })}
}

func (e Event) WithDestinationID(id EventDestinationID) Event {
	return Event{e.forceUpdate(func(m *EventPB) { m.DestinationId = id.String() })}
}

func (e Event) WithConnectionDestinationID(id ConnectionID) Event {
	return e.WithDestinationID(NewEventDestinationID(id))
}

func (e Event) WithTriggerDestinationID(id TriggerID) Event {
	return e.WithDestinationID(NewEventDestinationID(id))
}

func (e Event) DestinationID() EventDestinationID {
	return kittehs.Must1(ParseEventDestinationID(e.read().DestinationId))
}

func (e Event) Memo() map[string]string { return e.read().Memo }

func (e Event) Type() string { return e.read().EventType }
func (e Event) WithType(t string) Event {
	return Event{e.forceUpdate(func(m *EventPB) { m.EventType = t })}
}

func (e Event) ToValues() map[string]Value {
	if !e.IsValid() {
		return nil
	}

	return map[string]Value{
		"type": NewStringValue(e.m.EventType),
		"id":   NewStringValue(e.m.EventId),
		"data": kittehs.Must1(NewStructValue(
			NewStringValue("event_data"),
			kittehs.TransformMapValues(
				kittehs.FilterMapKeys(e.m.Data, IsValidSymbol),
				forceFromProto[Value],
			),
		)),
		"created_at": NewTimeValue(e.CreatedAt()),
	}
}

func (e Event) CreatedAt() time.Time { return e.read().CreatedAt.AsTime() }
func (e Event) Seq() uint64          { return e.read().Seq }

func (e Event) Data() map[string]Value {
	return kittehs.TransformMapValues(e.read().Data, forceFromProto[Value])
}

func (e Event) WithData(data map[string]Value) Event {
	return Event{e.forceUpdate(func(m *EventPB) { m.Data = kittehs.TransformMapValues(data, func(v Value) *ValuePB { return v.m }) })}
}

var eventFilterEnv = kittehs.Must1(cel.NewEnv(
	cel.Variable("data", cel.MapType(cel.StringType, cel.AnyType)),
	cel.Variable("event_type", cel.StringType),
))

func VerifyEventFilter(filter string) error {
	if filter == "" {
		return nil
	}

	if _, err := eventFilterEnv.Compile(filter); err != nil {
		return err.Err()
	}

	return nil
}

func eventFilterField(name string, expr string) error {
	if expr == "" {
		return nil
	}

	_, issues := eventFilterEnv.Compile(expr)
	if err := issues.Err(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}

	return nil
}

var matchUnwrapper = ValueWrapper{
	Preunwrap: func(v Value) (Value, error) {
		// Ignore functions.
		if v.IsFunction() {
			return InvalidValue, nil
		}
		return v, nil
	},
}

func (e Event) Matches(expr string) (bool, error) {
	if expr == "" {
		return true, nil
	}

	ast, issues := eventFilterEnv.Compile(expr)
	if err := issues.Err(); err != nil {
		return false, fmt.Errorf("compile: %w", err)
	}

	prg, err := eventFilterEnv.Program(ast)
	if err != nil {
		return false, fmt.Errorf("program: %w", err)
	}

	data, err := kittehs.TransformMapValuesError(e.Data(), matchUnwrapper.Unwrap)
	if err != nil {
		return false, fmt.Errorf("unwrap event: %w", err)
	}

	out, _, err := prg.Eval(map[string]any{"data": data, "event_type": e.Type()})
	if err != nil {
		return false, fmt.Errorf("program eval: %w", err)
	}

	b, err := out.ConvertToNative(reflect.TypeOf(true))
	if err != nil {
		return false, fmt.Errorf("expression result not bool: %w", err)
	}

	return b.(bool), nil
}
