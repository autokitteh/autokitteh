package sdktypes

import (
	"fmt"

	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
)

type CallFramePB = programv1.CallFrame

type CallFrame = *object[*CallFramePB]

var (
	CallFrameFromProto       = makeFromProto(validateCallFrame)
	StrictCallFrameFromProto = makeFromProto(strictValidateCallFrame)
	ToStrictCallFrame        = makeWithValidator(strictValidateCallFrame)
)

func strictValidateCallFrame(pb *programv1.CallFrame) error {
	return validateCallFrame(pb)
}

func validateCallFrame(pb *programv1.CallFrame) error {
	if _, err := CodeLocationFromProto(pb.Location); err != nil {
		return fmt.Errorf("location: %w", err)
	}

	return nil
}
