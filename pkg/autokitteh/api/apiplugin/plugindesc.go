package apiplugin

import (
	"google.golang.org/protobuf/proto"

	pbplugin "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/plugin"
)

type PluginDesc struct{ pb *pbplugin.PluginDesc }

func (p *PluginDesc) PB() *pbplugin.PluginDesc {
	if p == nil || p.pb == nil {
		return nil
	}

	return proto.Clone(p.pb).(*pbplugin.PluginDesc)
}

func (p *PluginDesc) Clone() *PluginDesc {
	if p == nil || p.pb == nil {
		return nil
	}

	return &PluginDesc{pb: p.PB()}
}

func (p *PluginDesc) Doc() string { return p.pb.Doc }

func MustPluginDescFromProto(pb *pbplugin.PluginDesc) *PluginDesc {
	p, err := PluginDescFromProto(pb)
	if err != nil {
		panic(err)
	}
	return p
}

func PluginDescFromProto(pb *pbplugin.PluginDesc) (*PluginDesc, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	return (&PluginDesc{pb: pb}).Clone(), nil
}

func NewPluginDesc(doc string, members []*PluginMemberDesc) (*PluginDesc, error) {
	pbmembers := make([]*pbplugin.PluginMemberDesc, len(members))
	for i, m := range members {
		pbmembers[i] = m.PB()
	}

	return PluginDescFromProto(&pbplugin.PluginDesc{
		Doc:     doc,
		Members: pbmembers,
	})
}

func MustNewPluginDesc(doc string, members []*PluginMemberDesc) *PluginDesc {
	p, err := NewPluginDesc(doc, members)
	if err != nil {
		panic(err)
	}
	return p
}
