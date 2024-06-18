package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
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
		valuesMapField("locals", m.Locals),
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

func NewCallFrame(name string, loc CodeLocation, locals map[string]Value) CallFrame {
	return kittehs.Must1(CallFrameFromProto(&CallFramePB{
		Name:     name,
		Location: loc.ToProto(),
		Locals:   kittehs.TransformMapValues(locals, ToProto),
	}))
}
