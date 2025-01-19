package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	varsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/vars/v1"
)

type Var struct{ object[*VarPB, VarTraits] }

func init() { registerObject[Var]() }

var InvalidVar Var

type VarPB = varsv1.Var

type VarTraits struct{ immutableObjectTrait }

func (VarTraits) Validate(m *VarPB) error {
	return errors.Join(
		nameField("name", m.Name),
		varScopeIDField(m.ScopeId),
	)
}

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

func (p Var) SetSecret(s bool) Var {
	return Var{p.forceUpdate(func(pb *VarPB) { pb.IsSecret = s })}
}

func (p Var) SetValue(v string) Var {
	return Var{p.forceUpdate(func(pb *VarPB) { pb.Value = v })}
}

func NewVar(n Symbol) Var {
	return kittehs.Must1(StrictVarFromProto(&VarPB{Name: n.String()}))
}
