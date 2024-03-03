package sdktypes

import (
	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
)

type ModuleVariable struct {
	object[*ModuleVariablePB, ModuleVariableTraits]
}

type ModuleVariablePB = programv1.Variable

type ModuleVariableTraits struct{}

func (ModuleVariableTraits) Validate(m *ModuleVariablePB) error       { return nil }
func (ModuleVariableTraits) StrictValidate(m *ModuleVariablePB) error { return nil }

func ModuleVariableFromProto(m *ModuleVariablePB) (ModuleVariable, error) {
	return FromProto[ModuleVariable](m)
}

func StrictModuleVariableFromProto(m *ModuleVariablePB) (ModuleVariable, error) {
	return Strict(ModuleVariableFromProto(m))
}
