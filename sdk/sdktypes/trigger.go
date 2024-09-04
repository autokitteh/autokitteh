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
		idField[EnvID]("env_id", m.EnvId),
		idField[TriggerID]("trigger_id", m.TriggerId),
		objectField[CodeLocation]("code_location", m.CodeLocation),
		symbolField("name", m.Name),
		idField[ConnectionID]("connection_id", m.ConnectionId),
		enumField[TriggerSourceType]("source_type", m.SourceType),
	)
}

func (TriggerTraits) StrictValidate(m *TriggerPB) error {
	var err error

	switch m.SourceType {
	case triggerv1.Trigger_SOURCE_TYPE_SCHEDULE:
		if m.Schedule == "" {
			err = errors.New("schedule is required for schedule trigger")
		}
	case triggerv1.Trigger_SOURCE_TYPE_CONNECTION:
		if m.ConnectionId == "" {
			err = errors.New("connection id is required for connection")
		}
	case triggerv1.Trigger_SOURCE_TYPE_WEBHOOK:
		// nop
	}

	return errors.Join(
		err,
		mandatory("name", m.Name),
		mandatory("env_id", m.EnvId),
		mandatory("source_type", m.SourceType),
	)
}

func TriggerFromProto(m *TriggerPB) (Trigger, error)       { return FromProto[Trigger](m) }
func StrictTriggerFromProto(m *TriggerPB) (Trigger, error) { return Strict(TriggerFromProto(m)) }

func (p Trigger) ID() TriggerID { return kittehs.Must1(ParseTriggerID(p.read().TriggerId)) }

func (p Trigger) WithNewID() Trigger { return p.WithID(NewTriggerID()) }

func (p Trigger) WithID(id TriggerID) Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) { m.TriggerId = id.String() })}
}

func (p Trigger) EnvID() EnvID { return kittehs.Must1(ParseEnvID(p.read().EnvId)) }
func (p Trigger) WithEnvID(id EnvID) Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) { m.EnvId = id.String() })}
}

func (p Trigger) Name() Symbol { return NewSymbol(p.read().Name) }
func (p Trigger) WithName(s Symbol) Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) { m.Name = s.String() })}
}

func (p Trigger) EventType() string { return p.read().EventType }

func (p Trigger) Schedule() string { return p.read().Schedule }
func (p Trigger) WithSchedule(expr string) Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) {
		m.Schedule = expr
		m.SourceType = triggerv1.Trigger_SOURCE_TYPE_SCHEDULE
	})}
}

func (p Trigger) WebhookSlug() string { return p.read().WebhookSlug }
func (p Trigger) WithWebhookSlug(slug string) Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) {
		m.WebhookSlug = slug
		if slug != "" {
			m.SourceType = triggerv1.Trigger_SOURCE_TYPE_WEBHOOK
		}
	})}
}

func (p Trigger) ConnectionID() ConnectionID {
	return kittehs.Must1(ParseConnectionID(p.read().ConnectionId))
}

func (p Trigger) WithConnectionID(id ConnectionID) Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) {
		m.ConnectionId = id.String()
		m.SourceType = triggerv1.Trigger_SOURCE_TYPE_CONNECTION
	})}
}

func (p Trigger) Filter() string { return p.read().Filter }
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

	return map[string]Value{"name": NewStringValue(p.read().Name)}
}

func (p Trigger) SourceType() TriggerSourceType {
	return kittehs.Must1(EnumFromProto[TriggerSourceType](p.read().SourceType))
}

func (p Trigger) WithSourceType(t TriggerSourceType) Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) { m.SourceType = t.ToProto() })}
}

func (p Trigger) WithWebhook() Trigger {
	return Trigger{p.forceUpdate(func(m *TriggerPB) {
		m.SourceType = triggerv1.Trigger_SOURCE_TYPE_WEBHOOK
	})}
}
