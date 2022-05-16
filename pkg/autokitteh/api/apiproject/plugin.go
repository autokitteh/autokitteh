package apiproject

import (
	"google.golang.org/protobuf/proto"

	pbproject "github.com/autokitteh/autokitteh/gen/proto/stubs/go/project"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiplugin"
)

type ProjectPlugin struct{ pb *pbproject.ProjectPlugin }

func (p *ProjectPlugin) PB() *pbproject.ProjectPlugin {
	if p == nil || p.pb == nil {
		return nil
	}

	return proto.Clone(p.pb).(*pbproject.ProjectPlugin)
}

func (p *ProjectPlugin) Clone() *ProjectPlugin { return &ProjectPlugin{pb: p.PB()} }

func (p *ProjectPlugin) ID() apiplugin.PluginID { return apiplugin.PluginID(p.pb.PluginId) }
func (p *ProjectPlugin) Enabled() bool          { return p.pb.Enabled }

func MustProjectPluginFromProto(pb *pbproject.ProjectPlugin) *ProjectPlugin {
	pl, err := ProjectPluginFromProto(pb)
	if err != nil {
		panic(err)
	}
	return pl
}

func ProjectPluginFromProto(pb *pbproject.ProjectPlugin) (*ProjectPlugin, error) {
	if pb == nil {
		return nil, nil
	}

	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&ProjectPlugin{pb: pb}).Clone(), nil
}

func MustNewProjectPlugin(id apiplugin.PluginID, enabled bool) *ProjectPlugin {
	pl, err := NewProjectPlugin(id, enabled)
	if err != nil {
		panic(err)
	}
	return pl
}

func NewProjectPlugin(id apiplugin.PluginID, enabled bool) (*ProjectPlugin, error) {
	return ProjectPluginFromProto(
		&pbproject.ProjectPlugin{
			PluginId: id.String(),
			Enabled:  enabled,
		},
	)
}
