package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	envsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1"
)

type EnvPB = envsv1.Env

type Env = *object[*EnvPB]

var (
	EnvFromProto       = makeFromProto(validateEnv)
	StrictEnvFromProto = makeFromProto(strictValidateEnv)
	ToStrictEnv        = makeWithValidator(strictValidateEnv)
)

func strictValidateEnv(pb *envsv1.Env) error {
	if err := ensureNotEmpty(pb.EnvId, pb.ProjectId, pb.Name); err != nil {
		return err
	}

	return validateEnv(pb)
}

func validateEnv(pb *envsv1.Env) error {
	if _, err := ParseEnvID(pb.EnvId); err != nil {
		return fmt.Errorf("env ID: %w", err)
	}

	if _, err := ParseProjectID(pb.ProjectId); err != nil {
		return fmt.Errorf("project ID: %w", err)
	}

	return nil
}

func EnvHasID(e Env) bool { return e.pb.EnvId != "" }

func GetEnvID(e Env) EnvID {
	if e == nil {
		return nil
	}

	return kittehs.Must1(ParseEnvID(e.pb.EnvId))
}

func GetEnvProjectID(e Env) ProjectID {
	if e == nil {
		return nil
	}

	return kittehs.Must1(ParseProjectID(e.pb.ProjectId))
}

func GetEnvName(e Env) Name { return kittehs.Must1(ParseName(e.pb.Name)) }
