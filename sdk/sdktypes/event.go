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
	if !e.IsValid() {
		return nil
	}

	return map[string]Value{
		"type":           NewStringValue(e.m.EventType),
		"id":             NewStringValue(e.m.EventId),
		"original_id":    NewStringValue(e.m.OriginalEventId),
		"integration_id": NewStringValue(e.m.IntegrationId),
		"data": kittehs.Must1(NewStructValue(
			NewStringValue("event_data"),
			kittehs.TransformMapValues(
				kittehs.FilterMapKeys(e.m.Data, IsValidSymbol),
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

var eventFilterEnv = kittehs.Must1(cel.NewEnv(
	cel.Variable("data", cel.MapType(cel.StringType, cel.AnyType)),
))

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

	data, err := kittehs.TransformMapValuesError(e.Data(), UnwrapValue)
	if err != nil {
		return false, fmt.Errorf("convert: %w", err)
	}

	out, _, err := prg.Eval(map[string]any{"data": data})
	if err != nil {
		return false, fmt.Errorf("eval: %w", err)
	}

	b, err := out.ConvertToNative(reflect.TypeOf(true))
	if err != nil {
		return false, fmt.Errorf("expression result is not a boolean: %w", err)
	}

	return b.(bool), nil
}
