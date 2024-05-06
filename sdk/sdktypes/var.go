package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	varsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/vars/v1"
)

type Var struct{ object[*VarPB, VarTraits] }

var InvalidVar Var

type VarPB = varsv1.Var

type VarTraits struct{}

func (VarTraits) Validate(m *VarPB) error       { return nameField("name", m.Name) }
func (VarTraits) StrictValidate(m *VarPB) error { return mandatory("name", m.Name) }

func VarFromProto(m *VarPB) (Var, error)       { return FromProto[Var](m) }
func StrictVarFromProto(m *VarPB) (Var, error) { return Strict(VarFromProto(m)) }

func (p Var) ScopeID() VarScopeID { return kittehs.Must1(ParseVarScopeID(p.read().ScopeId)) }
func (p Var) Name() Symbol        { return kittehs.Must1(ParseSymbol(p.read().Name)) }
func (p Var) Value() string       { return p.read().Value }
func (p Var) IsSecret() bool      { return p.read().IsSecret }

func (p Var) WithScopeID(id VarScopeID) Var {
	return Var{p.forceUpdate(func(pb *VarPB) { pb.ScopeId = id.String() })}
}

func NewVar(n Symbol, v string, isSecret bool) Var {
	return kittehs.Must1(StrictVarFromProto(&VarPB{Name: n.String(), Value: v, IsSecret: isSecret}))
}
