package apieventsrc

import (
	"google.golang.org/protobuf/proto"

	pbeventsrc "github.com/autokitteh/autokitteh/gen/proto/stubs/go/eventsrc"
)

type EventSourceSettingsPB = pbeventsrc.EventSourceSettings

type EventSourceSettings struct {
	pb *pbeventsrc.EventSourceSettings
}

func (a *EventSourceSettings) PB() *pbeventsrc.EventSourceSettings {
	if a == nil {
		return nil
	}
	return proto.Clone(a.pb).(*pbeventsrc.EventSourceSettings)
}

func (a *EventSourceSettings) Clone() *EventSourceSettings {
	return &EventSourceSettings{pb: a.PB()}
}

func (a *EventSourceSettings) prep() *EventSourceSettings {
	if a == nil || a.pb == nil {
		return &EventSourceSettings{pb: &pbeventsrc.EventSourceSettings{}}
	}

	return a.Clone()
}

func (a *EventSourceSettings) Enabled() bool {
	if a == nil || a.pb == nil {
		return false
	}
	return a.pb.Enabled
}

func (a *EventSourceSettings) SetEnabled(e bool) *EventSourceSettings {
	a = a.prep()
	a.pb.Enabled = e
	return a
}

func (a *EventSourceSettings) Types() []string {
	if a == nil || a.pb == nil {
		return nil
	}
	return a.pb.Types
}

func (a *EventSourceSettings) SetTypes(ts []string) *EventSourceSettings {
	a = a.prep()
	a.pb.Types = ts
	return a
}

func MustEventSourceSettingsFromProto(pb *pbeventsrc.EventSourceSettings) *EventSourceSettings {
	d, err := EventSourceSettingsFromProto(pb)
	if err != nil {
		panic(err)
	}
	return d
}

func EventSourceSettingsFromProto(pb *pbeventsrc.EventSourceSettings) (*EventSourceSettings, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&EventSourceSettings{pb: pb}).Clone(), nil
}
