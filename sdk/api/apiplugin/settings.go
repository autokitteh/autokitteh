package apiplugin

import (
	"google.golang.org/protobuf/proto"

	pbplugin "github.com/autokitteh/autokitteh/api/gen/stubs/go/plugin"
)

type PluginSettingsPB = pbplugin.PluginSettings

type PluginSettings struct{ pb *pbplugin.PluginSettings }

func (d *PluginSettings) PB() *pbplugin.PluginSettings {
	if d == nil {
		return nil
	}

	return proto.Clone(d.pb).(*pbplugin.PluginSettings)
}

func (d *PluginSettings) prep() *PluginSettings {
	if d == nil || d.pb == nil {
		return &PluginSettings{pb: &pbplugin.PluginSettings{}}
	}

	return d.Clone()
}

func (d *PluginSettings) Clone() *PluginSettings { return &PluginSettings{pb: d.PB()} }

func (p *PluginSettings) Port() uint16 {
	if p == nil || p.pb == nil {
		return 0
	}

	return uint16(p.pb.Port)
}

func (p *PluginSettings) SetPort(port uint16) *PluginSettings {
	p = p.prep()
	p.pb.Port = uint32(port)
	return p
}

func (p *PluginSettings) Address() string {
	if p == nil || p.pb == nil {
		return ""
	}

	return p.pb.Address
}

func (p *PluginSettings) SetAddress(a string) *PluginSettings {
	p = p.prep()
	p.pb.Address = a
	return p
}

func (p *PluginSettings) Exec() *PluginExecSettings {
	if p == nil || p.pb == nil {
		return nil
	}

	return &PluginExecSettings{pb: p.pb.Exec}
}

func (p *PluginSettings) SetExec(x *PluginExecSettings) *PluginSettings {
	p = p.prep()
	p.pb.Exec = x.PB()
	return p
}

func (p *PluginSettings) Enabled() bool {
	if p == nil || p.pb == nil {
		return false
	}

	return p.pb.Enabled
}

func (p *PluginSettings) SetEnabled(e bool) *PluginSettings {
	p = p.prep()
	p.pb.Enabled = e
	return p
}

func MustPluginSettingsFromProto(pb *pbplugin.PluginSettings) *PluginSettings {
	d, err := PluginSettingsFromProto(pb)
	if err != nil {
		panic(err)
	}

	return d
}

func PluginSettingsFromProto(pb *pbplugin.PluginSettings) (*PluginSettings, error) {
	if pb == nil {
		return nil, nil
	}

	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&PluginSettings{pb: pb}).Clone(), nil
}
