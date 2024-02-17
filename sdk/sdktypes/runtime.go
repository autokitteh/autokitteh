package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
)

type RuntimePB = runtimesv1.Runtime

type Runtime = *object[*RuntimePB]

var (
	RuntimeFromProto       = makeFromProto(validateRuntime)
	StrictRuntimeFromProto = makeFromProto(strictValidateRuntime)
	ToStrictRuntime        = makeWithValidator(strictValidateRuntime)
)

func strictValidateRuntime(pb *runtimesv1.Runtime) error {
	if err := ensureNotEmpty(pb.Name); err != nil {
		return err
	}

	return validateRuntime(pb)
}

func validateRuntime(pb *runtimesv1.Runtime) error {
	if _, err := ParseName(pb.Name); err != nil {
		return fmt.Errorf("name: %w", err)
	}

	return nil
}

func GetRuntimeName(r Runtime) Name               { return kittehs.Must1(ParseName(r.pb.Name)) }
func GetRuntimeFileExtensions(r Runtime) []string { return r.pb.FileExtensions }
