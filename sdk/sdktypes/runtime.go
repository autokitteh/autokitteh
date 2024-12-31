package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
)

type Runtime struct {
	object[*RuntimePB, RuntimeTraits]
}

type RuntimePB = runtimev1.Runtime

type RuntimeTraits struct{ immutableObjectTrait }

func (RuntimeTraits) Validate(m *RuntimePB) error {
	return nameField("name", m.Name)
}

func (RuntimeTraits) StrictValidate(m *RuntimePB) error {
	return errors.Join(
		mandatory("name", m.Name),
		mandatorySlice("file_extensions", m.FileExtensions),
	)
}

func RuntimeFromProto(m *RuntimePB) (Runtime, error)       { return FromProto[Runtime](m) }
func StrictRuntimeFromProto(m *RuntimePB) (Runtime, error) { return Strict(RuntimeFromProto(m)) }

func (r Runtime) Name() Symbol             { return kittehs.Must1(ParseSymbol(r.read().Name)) }
func (r Runtime) FileExtensions() []string { return r.read().FileExtensions }
