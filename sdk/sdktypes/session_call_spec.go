package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionCallSpec struct {
	object[*SessionCallSpecPB, SessionCallSpecTraits]
}

var InvalidSessionCallSpec SessionCallSpec

type SessionCallSpecPB = sessionv1.Call_Spec

type SessionCallSpecTraits struct{}

func (SessionCallSpecTraits) Validate(m *SessionCallSpecPB) error {
	return errors.Join(
		objectField[Value]("function", m.Function),
		valuesSliceField("args", m.Args),
		valuesMapField("kwargs", m.Kwargs),
	)
}

func (SessionCallSpecTraits) StrictValidate(m *SessionCallSpecPB) error {
	return mandatory("function", m.Function)
}

func SessionCallSpecFromProto(m *SessionCallSpecPB) (SessionCallSpec, error) {
	return FromProto[SessionCallSpec](m)
}

func StrictSessionCallSpecFromProto(m *SessionCallSpecPB) (SessionCallSpec, error) {
	return Strict(SessionCallSpecFromProto(m))
}

func (p SessionCallSpec) Seq() uint32 { return p.read().Seq }

func (p SessionCallSpec) Data() (Value, []Value, map[string]Value) {
	pb := p.read()

	return forceFromProto[Value](pb.Function),
		kittehs.Transform(pb.Args, forceFromProto[Value]),
		kittehs.TransformMapValues(pb.Kwargs, forceFromProto[Value])
}

func NewSessionCallSpec(function Value, args []Value, kwargs map[string]Value, seq uint32) SessionCallSpec {
	return forceFromProto[SessionCallSpec](&SessionCallSpecPB{
		Seq:      seq,
		Function: function.ToProto(),
		Args:     kittehs.Transform(args, ToProto),
		Kwargs:   kittehs.TransformMapValues(kwargs, ToProto),
	})
}
