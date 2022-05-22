package apiproject

import (
	"google.golang.org/protobuf/proto"

	pbproject "github.com/autokitteh/autokitteh/api/gen/stubs/go/project"

	"github.com/autokitteh/autokitteh/sdk/api/apiplugin"
	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
)

type ProjectSettingsPB = pbproject.ProjectSettings

type ProjectSettings struct{ pb *pbproject.ProjectSettings }

func (d *ProjectSettings) PB() *pbproject.ProjectSettings {
	if d == nil {
		return nil
	}

	return proto.Clone(d.pb).(*pbproject.ProjectSettings)
}

func (d *ProjectSettings) prep() *ProjectSettings {
	if d == nil || d.pb == nil {
		return &ProjectSettings{pb: &pbproject.ProjectSettings{}}
	}

	return d.Clone()
}

func (d *ProjectSettings) Clone() *ProjectSettings { return &ProjectSettings{pb: d.PB()} }

func (d *ProjectSettings) Memo() map[string]string {
	if d == nil || d.pb == nil {
		return nil
	}
	return d.pb.Memo
}

func (p *ProjectSettings) SetMemo(memo map[string]string) *ProjectSettings {
	p = p.prep()
	p.pb.Memo = memo
	return p
}

func (p *ProjectSettings) Enabled() bool {
	if p == nil || p.pb == nil {
		return false
	}

	return p.pb.Enabled
}

func (p *ProjectSettings) SetEnabled(e bool) *ProjectSettings {
	p = p.prep()
	p.pb.Enabled = e
	return p
}

func (d *ProjectSettings) Name() string {
	if d == nil || d.pb == nil {
		return ""
	}

	return d.pb.Name
}

func (p *ProjectSettings) SetName(n string) *ProjectSettings {
	p = p.prep()
	p.pb.Name = n
	return p
}

func (p *ProjectSettings) MainPath() *apiprogram.Path {
	if p == nil || p.pb == nil {
		return nil
	}

	return apiprogram.MustPathFromProto(p.pb.MainPath)
}

func (p *ProjectSettings) SetMainPath(path *apiprogram.Path) *ProjectSettings {
	p = p.prep()
	p.pb.MainPath = path.PB()
	return p
}

func (p *ProjectSettings) AddPlugin(pl *ProjectPlugin) *ProjectSettings {
	p = p.prep()

	p.pb.Plugins = append(p.pb.Plugins, pl.PB())

	return p
}

func (p *ProjectSettings) SetPlugins(pls []*ProjectPlugin) *ProjectSettings {
	p = p.prep()

	p.pb.Plugins = make([]*pbproject.ProjectPlugin, len(pls))
	for i, pl := range pls {
		p.pb.Plugins[i] = pl.PB()
	}

	return p
}

func (p *ProjectSettings) Plugins() []*ProjectPlugin {
	if p == nil || p.pb == nil {
		return nil
	}

	pls := make([]*ProjectPlugin, len(p.pb.Plugins))
	for i, pl := range p.pb.Plugins {
		pls[i] = MustProjectPluginFromProto(pl)
	}
	return pls
}

func (p *ProjectSettings) Plugin(id apiplugin.PluginID) *ProjectPlugin {
	if p == nil || p.pb == nil {
		return nil
	}

	for _, pl := range p.pb.Plugins {
		if pl.PluginId == string(id) {
			return MustProjectPluginFromProto(pl)
		}
	}

	return nil
}

func (p *ProjectSettings) SetPredecls(m map[string]*apivalues.Value) *ProjectSettings {
	p = p.prep()
	p.pb.Predecls = apivalues.StringValueMapToProto(m)
	return p
}

func (p *ProjectSettings) Predecls() map[string]*apivalues.Value {
	if p == nil || p.pb == nil {
		return nil
	}

	return apivalues.MustStringValueMapFromProto(p.pb.Predecls)
}

func MustProjectSettingsFromProto(pb *pbproject.ProjectSettings) *ProjectSettings {
	d, err := ProjectSettingsFromProto(pb)
	if err != nil {
		panic(err)
	}

	return d
}

func ProjectSettingsFromProto(pb *pbproject.ProjectSettings) (*ProjectSettings, error) {
	if pb == nil {
		return nil, nil
	}

	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&ProjectSettings{pb: pb}).Clone(), nil
}
