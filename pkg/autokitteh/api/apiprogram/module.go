package apiprogram

import (
	"sort"

	"google.golang.org/protobuf/proto"

	pbprogram "github.com/autokitteh/autokitteh/api/gen/stubs/go/program"
)

type Module struct{ pb *pbprogram.Module }

func (m *Module) PB() *pbprogram.Module { return proto.Clone(m.pb).(*pbprogram.Module) }
func (m *Module) Clone() *Module        { return &Module{pb: m.PB()} }

func (m *Module) Validate() error { return m.pb.Validate() }

func ModuleFromProto(pb *pbprogram.Module) (*Module, error) {
	if pb == nil {
		return nil, nil
	}

	if err := pb.Validate(); err != nil {
		return nil, err
	}

	sort.Strings(pb.Predecls)

	return (&Module{pb: pb}).Clone(), nil
}

func MustModuleFromProto(pb *pbprogram.Module) *Module {
	m, err := ModuleFromProto(pb)
	if err != nil {
		panic(err)
	}

	return m
}

func (m *Module) CompiledCode() []byte    { return m.pb.CompiledCode }
func (m *Module) CompilerVersion() string { return m.pb.CompilerVersion }
func (m *Module) Lang() string            { return m.pb.Lang }

func (m *Module) SourcePath() *Path {
	p, err := PathFromProto(m.pb.SourcePath)
	if err != nil {
		panic(err)
	}

	return p
}

func (m *Module) SetSourcePath(p *Path) *Module {
	m = m.Clone()
	m.pb.SourcePath = p.PB()
	return m
}

func NewModule(lang string, predecls []string, compilerVersion string, path *Path, compiled []byte) (*Module, error) {
	mod, err := ModuleFromProto(
		&pbprogram.Module{
			Lang:            lang,
			Predecls:        predecls,
			CompilerVersion: compilerVersion,
			SourcePath:      path.PB(),
			CompiledCode:    compiled,
		},
	)

	if err != nil {
		return nil, err
	}

	sort.Strings(mod.pb.Predecls)

	return mod, nil
}
