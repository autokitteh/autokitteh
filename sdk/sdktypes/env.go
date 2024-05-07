package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	envv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1"
)

type Env struct{ object[*EnvPB, EnvTraits] }

var InvalidEnv Env

type EnvPB = envv1.Env

type EnvTraits struct{}

func (EnvTraits) Validate(m *EnvPB) error {
	return errors.Join(
		idField[EnvID]("env_id", m.EnvId),
		nameField("name", m.Name),
		idField[ProjectID]("project_id", m.ProjectId),
	)
}

func (EnvTraits) StrictValidate(m *EnvPB) error {
	return errors.Join(
		mandatory("env_id", m.EnvId),
		mandatory("name", m.Name),
		mandatory("project_id", m.ProjectId),
	)
}

func EnvFromProto(m *EnvPB) (Env, error)       { return FromProto[Env](m) }
func StrictEnvFromProto(m *EnvPB) (Env, error) { return Strict(EnvFromProto(m)) }

func (p Env) ID() EnvID            { return kittehs.Must1(ParseEnvID(p.read().EnvId)) }
func (p Env) ProjectID() ProjectID { return kittehs.Must1(ParseProjectID(p.read().ProjectId)) }
func (p Env) Name() Symbol         { return NewSymbol(p.read().Name) }

func NewEnv() Env {
	return kittehs.Must1(EnvFromProto(&EnvPB{}))
}

func (p Env) WithName(name Symbol) Env {
	return Env{p.forceUpdate(func(pb *EnvPB) { pb.Name = name.String() })}
}

func (p Env) WithNewID() Env { return p.WithID(NewEnvID()) }

func (p Env) WithID(id EnvID) Env {
	return Env{p.forceUpdate(func(pb *EnvPB) { pb.EnvId = id.String() })}
}

func (p Env) WithProjectID(id ProjectID) Env {
	return Env{p.forceUpdate(func(pb *EnvPB) { pb.ProjectId = id.String() })}
}
