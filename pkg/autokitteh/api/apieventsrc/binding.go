package apieventsrc

import (
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbeventsrc "github.com/autokitteh/autokitteh/api/gen/stubs/go/eventsrc"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
)

type EventSourceProjectBindingPB = pbeventsrc.EventSourceProjectBinding

type EventSourceProjectBinding struct {
	pb *pbeventsrc.EventSourceProjectBinding
}

func (b *EventSourceProjectBinding) PB() *pbeventsrc.EventSourceProjectBinding {
	if b == nil || b.pb == nil {
		return nil
	}

	return proto.Clone(b.pb).(*pbeventsrc.EventSourceProjectBinding)
}

func (b *EventSourceProjectBinding) Clone() *EventSourceProjectBinding {
	return &EventSourceProjectBinding{pb: b.PB()}
}

func (b *EventSourceProjectBinding) Name() string {
	return b.pb.Name
}

func (b *EventSourceProjectBinding) Approved() bool {
	return b.pb.Approved
}

func (b *EventSourceProjectBinding) EventSourceID() EventSourceID {
	return EventSourceID(b.pb.SrcId)
}

func (b *EventSourceProjectBinding) AssociationToken() string {
	return b.pb.AssociationToken
}

func (b *EventSourceProjectBinding) SourceConfig() string {
	return b.pb.SourceConfig
}

func (b *EventSourceProjectBinding) ProjectID() apiproject.ProjectID {
	return apiproject.ProjectID(b.pb.ProjectId)
}

func (b *EventSourceProjectBinding) Settings() *EventSourceProjectBindingSettings {
	return MustEventSourceProjectBindingSettingsFromProto(b.pb.Settings)
}

func (b *EventSourceProjectBinding) WithoutTimes() *EventSourceProjectBinding {
	b = b.Clone()
	b.pb.CreatedAt = nil
	b.pb.UpdatedAt = nil
	return b
}

func EventSourceProjectBindingFromProto(pb *pbeventsrc.EventSourceProjectBinding) (*EventSourceProjectBinding, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&EventSourceProjectBinding{pb: pb}).Clone(), nil
}

func MustNewEventSourceProjectBinding(
	srcid EventSourceID,
	pid apiproject.ProjectID,
	name string,
	assoc, cfg string,
	approved bool,
	settings *EventSourceProjectBindingSettings,
	createdAt time.Time,
	updatedAt *time.Time,
) *EventSourceProjectBinding {
	b, err := NewEventSourceProjectBinding(srcid, pid, name, assoc, cfg, approved, settings, createdAt, updatedAt)
	if err != nil {
		panic(err)
	}
	return b
}

func NewEventSourceProjectBinding(
	srcid EventSourceID,
	pid apiproject.ProjectID,
	name string,
	assoc, cfg string,
	approved bool,
	settings *EventSourceProjectBindingSettings,
	createdAt time.Time,
	updatedAt *time.Time,
) (*EventSourceProjectBinding, error) {
	var upd *timestamppb.Timestamp
	if updatedAt != nil {
		upd = timestamppb.New(*updatedAt)
	}

	return &EventSourceProjectBinding{
		pb: &pbeventsrc.EventSourceProjectBinding{
			SrcId:            srcid.String(),
			ProjectId:        pid.String(),
			Name:             name,
			AssociationToken: assoc,
			SourceConfig:     cfg,
			Approved:         approved,
			Settings:         settings.PB(),
			CreatedAt:        timestamppb.New(createdAt),
			UpdatedAt:        upd,
		},
	}, nil
}
