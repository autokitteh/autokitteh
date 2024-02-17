package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	envsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1"
)

type EnvVarPB = envsv1.EnvVar

type EnvVar = *object[*EnvVarPB]

var (
	EnvVarFromProto       = makeFromProto(validateEnvVar)
	StrictEnvVarFromProto = makeFromProto(strictValidateEnvVar)
	ToStrictEnvVar        = makeWithValidator(strictValidateEnvVar)
)

func strictValidateEnvVar(pb *envsv1.EnvVar) error {
	if err := ensureNotEmpty(pb.EnvId, pb.Name); err != nil {
		return err
	}

	return validateEnvVar(pb)
}

func validateEnvVar(pb *envsv1.EnvVar) error {
	if _, err := ParseEnvID(pb.EnvId); err != nil {
		return fmt.Errorf("project id: %w", err)
	}

	if _, err := ParseSymbol(pb.Name); err != nil {
		return fmt.Errorf("name: %w", err)
	}

	return nil
}

func GetEnvVarEnvID(e EnvVar) EnvID {
	if e == nil {
		return nil
	}

	return kittehs.Must1(ParseEnvID(e.pb.EnvId))
}

func GetEnvVarName(e EnvVar) Symbol  { return kittehs.Must1(ParseSymbol(e.pb.Name)) }
func GetEnvVarValue(e EnvVar) string { return e.pb.Value }
func IsEnvVarSecret(e EnvVar) bool   { return e.pb.IsSecret }
