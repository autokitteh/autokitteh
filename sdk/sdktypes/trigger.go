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
		idField[TriggerID]("trigger_id", m.TriggerId),
		idField[ConnectionID]("connection_id", m.ConnectionId),
		idField[EnvID]("env_id", m.EnvId),
		objectField[CodeLocation]("code_location", m.CodeLocation),
		eventFilterField("filter", m.Filter),
	)
}

func (TriggerTraits) StrictValidate(m *TriggerPB) error {
	return errors.Join(
		mandatory("env_id", m.EnvId),
		mandatory("connection_id", m.ConnectionId),
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
func (p Trigger) EnvID() EnvID      { return kittehs.Must1(ParseEnvID(p.read().EnvId)) }
func (p Trigger) EventType() string { return p.read().EventType }
func (p Trigger) Filter() string    { return p.read().Filter }
func (p Trigger) CodeLocation() CodeLocation {
	return forceFromProto[CodeLocation](p.read().CodeLocation)
}
