package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	triggerv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/triggers/v1"
)

type Trigger struct {
	object[*TriggerPB, TriggerTraits]
}

var InvalidTrigger Trigger

type TriggerPB = triggerv1.Trigger

type TriggerTraits struct{}

func (TriggerTraits) Validate(m *TriggerPB) error {
	return errors.Join(
		eventFilterField("filter", m.Filter),
		idField[ConnectionID]("connection_id", m.ConnectionId),
		idField[EnvID]("env_id", m.EnvId),
		idField[TriggerID]("trigger_id", m.TriggerId),
		objectField[CodeLocation]("code_location", m.CodeLocation),
		valuesMapField("data", m.Data),
		symbolField("name", m.Name),
	)
}

func (TriggerTraits) StrictValidate(m *TriggerPB) error {
	return errors.Join(
		mandatory("env_id", m.EnvId),
		mandatory("name", m.Name),
	)
}

func TriggerFromProto(m *TriggerPB) (Trigger, error)       { return FromProto[Trigger](m) }
func StrictTriggerFromProto(m *TriggerPB) (Trigger, error) { return Strict(TriggerFromProto(m)) }

func (p Trigger) ID() TriggerID { return kittehs.Must1(ParseTriggerID(p.read().TriggerId)) }

func (p Trigger) WithNewID() Trigger { return p.WithID(NewTriggerID()) }

func (p Trigger) WithID(id TriggerID) Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) { m.TriggerId = id.String() })}
}

func (p Trigger) WithEnvID(id EnvID) Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) { m.EnvId = id.String() })}
}

func (p Trigger) WithConnectionID(id ConnectionID) Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) { m.ConnectionId = id.String() })}
}

func (p Trigger) ConnectionID() ConnectionID {
	return kittehs.Must1(ParseConnectionID(p.read().ConnectionId))
}

func (p Trigger) Data() map[string]Value {
	return kittehs.TransformMapValues(p.read().Data, forceFromProto[Value])
}

func (p Trigger) Name() Symbol { return NewSymbol(p.read().Name) }
func (p Trigger) WithName(s Symbol) Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) { m.Name = s.String() })}
}

func (p Trigger) EnvID() EnvID      { return kittehs.Must1(ParseEnvID(p.read().EnvId)) }
func (p Trigger) EventType() string { return p.read().EventType }
func (p Trigger) Filter() string    { return p.read().Filter }

func (p Trigger) WithFilter(f string) Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) { m.Filter = f })}
}

func (p Trigger) CodeLocation() CodeLocation {
	return forceFromProto[CodeLocation](p.read().CodeLocation)
}

func (p Trigger) ToValues() map[string]Value {
	if !p.IsValid() {
		return nil
	}

	return map[string]Value{
		"name": NewStringValue(p.read().Name),
		"data": kittehs.Must1(NewStructValue(
			NewStringValue("trigger_data"),
			kittehs.TransformMapValues(
				p.read().Data,
				forceFromProto[Value],
			),
		)),
	}
}

func (p Trigger) WithUpdatedData(key string, val Value) Trigger {
	data := p.read().Data
	if val.IsValid() {
		data[key] = val.ToProto()
	} else {
		delete(data, key)
	}
	return Trigger{p.forceUpdate(func(m *TriggerPB) { m.Data = data })}
}
