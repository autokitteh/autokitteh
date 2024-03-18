package sdktypes

import (
	"errors"

	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
)

type CallFrame struct {
	object[*CallFramePB, CallFrameTraits]
}

type CallFramePB = programv1.CallFrame

type CallFrameTraits struct{}

func (CallFrameTraits) Validate(m *CallFramePB) error {
	// No need to validate name, as it is a freeform string.
	return errors.Join(
		objectField[CodeLocation]("location", m.Location),
	)
}

func (CallFrameTraits) StrictValidate(m *CallFramePB) error {
	return nonzeroMessage(m)
}

func (f CallFrame) Name() string { return f.read().Name }

func (f CallFrame) Location() CodeLocation { return forceFromProto[CodeLocation](f.read().Location) }

func CallFrameFromProto(m *CallFramePB) (CallFrame, error) {
	return FromProto[CallFrame](m)
}
