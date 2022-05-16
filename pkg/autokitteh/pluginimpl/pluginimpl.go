package pluginimpl

import (
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiplugin"
)

type Plugin struct {
	ID      string // can be either plugin name or full id.
	Doc     string
	Members map[string]*PluginMember
}

func (p *Plugin) Desc() *apiplugin.PluginDesc {
	members := make([]*apiplugin.PluginMemberDesc, 0, len(p.Members))

	for k, v := range p.Members {
		members = append(members, apiplugin.MustNewPluginMemberDesc(k, v.Doc))
	}

	return apiplugin.MustNewPluginDesc(p.Doc, members)
}
