package apieventsrc

import (
	"google.golang.org/protobuf/proto"

	pbeventsrc "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/eventsrc"
)

type EventSourceProjectBindingSettingsPB = pbeventsrc.EventSourceProjectBindingSettings

type EventSourceProjectBindingSettings struct {
	pb *pbeventsrc.EventSourceProjectBindingSettings
}

func (a *EventSourceProjectBindingSettings) PB() *pbeventsrc.EventSourceProjectBindingSettings {
	if a == nil {
		return nil
	}
	return proto.Clone(a.pb).(*pbeventsrc.EventSourceProjectBindingSettings)
}

func (a *EventSourceProjectBindingSettings) Clone() *EventSourceProjectBindingSettings {
	return &EventSourceProjectBindingSettings{pb: a.PB()}
}

func (a *EventSourceProjectBindingSettings) prep() *EventSourceProjectBindingSettings {
	if a == nil || a.pb == nil {
		return &EventSourceProjectBindingSettings{pb: &pbeventsrc.EventSourceProjectBindingSettings{}}
	}

	return a.Clone()
}

func (a *EventSourceProjectBindingSettings) Enabled() bool {
	if a == nil || a.pb == nil {
		return false
	}
	return a.pb.Enabled
}

func (a *EventSourceProjectBindingSettings) SetEnabled(e bool) *EventSourceProjectBindingSettings {
	a = a.prep()
	a.pb.Enabled = e
	return a
}

func MustEventSourceProjectBindingSettingsFromProto(pb *pbeventsrc.EventSourceProjectBindingSettings) *EventSourceProjectBindingSettings {
	d, err := EventSourceProjectBindingSettingsFromProto(pb)
	if err != nil {
		panic(err)
	}
	return d
}

func EventSourceProjectBindingSettingsFromProto(pb *pbeventsrc.EventSourceProjectBindingSettings) (*EventSourceProjectBindingSettings, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&EventSourceProjectBindingSettings{pb: pb}).Clone(), nil
}
