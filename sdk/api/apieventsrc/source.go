package apieventsrc

import (
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbeventsrc "github.com/autokitteh/autokitteh/api/gen/stubs/go/eventsrc"
)

type EventSourcePB = pbeventsrc.EventSource

type EventSource struct{ pb *pbeventsrc.EventSource }

func (s *EventSource) PB() *pbeventsrc.EventSource {
	if s == nil || s.pb == nil {
		return nil
	}

	return proto.Clone(s.pb).(*pbeventsrc.EventSource)
}

func (s *EventSource) ID() EventSourceID { return EventSourceID(s.pb.Id) }

func (s *EventSource) Settings() *EventSourceSettings {
	return MustEventSourceSettingsFromProto(s.pb.Settings)
}

func (s *EventSource) Clone() *EventSource { return &EventSource{pb: s.PB()} }

func EventSourceFromProto(pb *pbeventsrc.EventSource) (*EventSource, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&EventSource{pb: pb}).Clone(), nil
}

func NewEventSource(
	id EventSourceID,
	settings *EventSourceSettings,
	createdAt time.Time,
	updatedAt *time.Time,
) (*EventSource, error) {
	var upd *timestamppb.Timestamp
	if updatedAt != nil {
		upd = timestamppb.New(*updatedAt)
	}

	return &EventSource{
		pb: &pbeventsrc.EventSource{
			Id:        id.String(),
			Settings:  settings.PB(),
			CreatedAt: timestamppb.New(createdAt),
			UpdatedAt: upd,
		},
	}, nil
}
