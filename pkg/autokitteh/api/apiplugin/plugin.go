package apiplugin

import (
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbplugin "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/plugin"
)

type PluginPB = pbplugin.Plugin

type Plugin struct{ pb *pbplugin.Plugin }

func (p *Plugin) PB() *pbplugin.Plugin {
	if p == nil || p.pb == nil {
		return nil
	}

	return proto.Clone(p.pb).(*pbplugin.Plugin)
}

func (p *Plugin) Clone() *Plugin { return &Plugin{pb: p.PB()} }

func (p *Plugin) ID() PluginID { return PluginID(p.pb.Id) }

func (p *Plugin) Settings() *PluginSettings { return MustPluginSettingsFromProto(p.pb.Settings) }

func PluginFromProto(pb *pbplugin.Plugin) (*Plugin, error) {
	if pb == nil {
		return nil, nil
	}

	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&Plugin{pb: pb}).Clone(), nil
}

func NewPlugin(id PluginID, d *PluginSettings, createdAt time.Time, updatedAt *time.Time) (*Plugin, error) {
	var pbupdatedat *timestamppb.Timestamp
	if updatedAt != nil {
		pbupdatedat = timestamppb.New(*updatedAt)
	}

	return PluginFromProto(
		&pbplugin.Plugin{
			Id:        id.String(),
			Settings:  d.PB(),
			CreatedAt: timestamppb.New(createdAt),
			UpdatedAt: pbupdatedat,
		},
	)
}
