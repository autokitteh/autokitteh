package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	envv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1"
)

type EnvVar struct {
	object[*EnvVarPB, EnvVarTraits]
}

type EnvVarPB = envv1.EnvVar

type EnvVarTraits struct{}

func (EnvVarTraits) Validate(m *EnvVarPB) error {
	return errors.Join(
		idField[EnvID]("env_id", m.EnvId),
		nameField("name", m.Name),
	)
}

func (EnvVarTraits) StrictValidate(m *EnvVarPB) error {
	return errors.Join(
		mandatory("env_id", m.EnvId),
		mandatory("name", m.Name),
	)
}

func EnvVarFromProto(m *EnvVarPB) (EnvVar, error)       { return FromProto[EnvVar](m) }
func StrictEnvVarFromProto(m *EnvVarPB) (EnvVar, error) { return Strict(EnvVarFromProto(m)) }

func (p EnvVar) EnvID() EnvID   { return kittehs.Must1(ParseEnvID(p.read().EnvId)) }
func (p EnvVar) Symbol() Symbol { return kittehs.Must1(ParseSymbol(p.read().Name)) }
func (p EnvVar) Value() string  { return p.read().Value }
func (p EnvVar) IsSecret() bool {
	return p.read().IsSecret
}

func (p EnvVar) WithEnvID(id EnvID) EnvVar {
	return EnvVar{p.forceUpdate(func(pb *EnvVarPB) { pb.EnvId = id.String() })}
}
