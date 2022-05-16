package apiplugin

import (
	"google.golang.org/protobuf/proto"

	pbplugin "github.com/autokitteh/autokitteh/gen/proto/stubs/go/plugin"
)

type PluginMemberDesc struct{ pb *pbplugin.PluginMemberDesc }

func (p *PluginMemberDesc) PB() *pbplugin.PluginMemberDesc {
	if p == nil || p.pb == nil {
		return nil
	}

	return proto.Clone(p.pb).(*pbplugin.PluginMemberDesc)
}

func (p *PluginMemberDesc) Clone() *PluginMemberDesc {
	if p == nil || p.pb == nil {
		return nil
	}

	return &PluginMemberDesc{pb: p.PB()}
}

func (p *PluginMemberDesc) Name() string { return p.pb.Name }
func (p *PluginMemberDesc) Doc() string  { return p.pb.Doc }

func MustPluginMemberDescFromProto(pb *pbplugin.PluginMemberDesc) *PluginMemberDesc {
	p, err := PluginMemberDescFromProto(pb)
	if err != nil {
		panic(err)
	}
	return p
}

func PluginMemberDescFromProto(pb *pbplugin.PluginMemberDesc) (*PluginMemberDesc, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	return (&PluginMemberDesc{pb: pb}).Clone(), nil
}

func NewPluginMemberDesc(name, doc string) (*PluginMemberDesc, error) {
	return PluginMemberDescFromProto(&pbplugin.PluginMemberDesc{
		Name: name,
		Doc:  doc,
	})
}

func MustNewPluginMemberDesc(name, doc string) *PluginMemberDesc {
	p, err := NewPluginMemberDesc(name, doc)
	if err != nil {
		panic(err)
	}
	return p
}
