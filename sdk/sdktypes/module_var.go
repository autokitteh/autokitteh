package sdktypes

import (
	modulev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/module/v1"
)

type ModuleVariable struct {
	object[*ModuleVariablePB, ModuleVariableTraits]
}

type ModuleVariablePB = modulev1.Variable

type ModuleVariableTraits struct{ immutableObjectTrait }

func (ModuleVariableTraits) Validate(m *ModuleVariablePB) error       { return nil }
func (ModuleVariableTraits) StrictValidate(m *ModuleVariablePB) error { return nil }

func ModuleVariableFromProto(m *ModuleVariablePB) (ModuleVariable, error) {
	return FromProto[ModuleVariable](m)
}

func StrictModuleVariableFromProto(m *ModuleVariablePB) (ModuleVariable, error) {
	return Strict(ModuleVariableFromProto(m))
}
