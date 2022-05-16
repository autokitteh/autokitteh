package apiplugin

import (
	"google.golang.org/protobuf/proto"

	pbplugin "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/plugin"
)

type PluginExecSettingsPB = pbplugin.PluginExecutionSettings

type PluginExecSettings struct {
	pb *pbplugin.PluginExecutionSettings
}

func (x *PluginExecSettings) PB() *pbplugin.PluginExecutionSettings {
	if x == nil {
		return nil
	}

	return proto.Clone(x.pb).(*pbplugin.PluginExecutionSettings)
}

func (x *PluginExecSettings) Clone() *PluginExecSettings { return &PluginExecSettings{pb: x.PB()} }

func (x *PluginExecSettings) Name() string {
	if x == nil || x.pb == nil {
		return ""
	}

	return x.pb.Name
}

func (p *PluginExecSettings) SetName(name string) *PluginExecSettings {
	p = p.prep()
	p.pb.Name = name
	return p
}

func (d *PluginExecSettings) prep() *PluginExecSettings {
	if d == nil || d.pb == nil {
		return &PluginExecSettings{pb: &pbplugin.PluginExecutionSettings{}}
	}

	return d.Clone()
}
