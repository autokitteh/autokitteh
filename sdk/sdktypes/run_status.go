package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
)

type RunStatusPB = runtimesv1.RunStatus

type RunStatus = *object[*RunStatusPB]

var (
	RunStatusFromProto       = makeFromProto(validateRunStatus)
	StrictRunStatusFromProto = makeFromProto(strictValidateRunStatus)
	ToStrictRunStatus        = makeWithValidator(strictValidateRunStatus)
)

func strictValidateRunStatus(pb *RunStatusPB) error {
	if pb.States == nil {
		return fmt.Errorf("nil state")
	}

	return validateRunStatus(pb)
}

func validateRunStatus(pb *RunStatusPB) error {
	// TODO
	return nil
}

func NewRunStatus(v Object) RunStatus {
	return kittehs.Must1(RunStatusFromProto(newRunStatusPB(v)))
}

func GetRunState(v RunStatus) Object { return kittehs.Must1(getRunStatus(v.pb)) }
