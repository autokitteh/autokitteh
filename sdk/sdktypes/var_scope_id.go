package sdktypes

import (
	"go.jetpack.io/typeid"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type VarScopeID struct{ id[typeid.AnyPrefix] }

var InvalidVarScopeID VarScopeID

type concreteVarScopeID interface {
	EnvID | ConnectionID
	ID
}

func NewVarScopeID[T concreteVarScopeID](in T) VarScopeID {
	parsed := kittehs.Must1(ParseID[id[typeid.AnyPrefix]](in.String()))
	return VarScopeID{parsed}
}

func ParseVarScopeID(s string) (VarScopeID, error) {
	if s == "" {
		return InvalidVarScopeID, nil
	}

	parsed, err := ParseID[id[typeid.AnyPrefix]](s)
	if err != nil {
		return InvalidVarScopeID, err
	}

	switch parsed.Kind() {
	case envIDKind, connectionIDKind:
		return VarScopeID{parsed}, nil
	default:
		return InvalidVarScopeID, sdkerrors.NewInvalidArgumentError("invalid executor id")
	}
}

func (e VarScopeID) ToEnvID() EnvID               { id, _ := ParseEnvID(e.String()); return id }
func (e VarScopeID) ToConnectionID() ConnectionID { id, _ := ParseConnectionID(e.String()); return id }

func (e VarScopeID) IsEnvID() bool        { return e.Kind() == envIDKind }
func (e VarScopeID) IsConnectionID() bool { return e.Kind() == connectionIDKind }

func (e VarScopeID) AsID() ID { return e }
