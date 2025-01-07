package sdktypes

import (
	"strings"

	modulev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/module/v1"
)

type ModuleFunctionField struct {
	object[*ModuleFunctionFieldPB, ModuleFunctionFieldTraits]
}

type ModuleFunctionFieldPB = modulev1.FunctionField

type ModuleFunctionFieldTraits struct{ immutableObjectTrait }

func (ModuleFunctionFieldTraits) Validate(m *ModuleFunctionFieldPB) error {
	return nameField("name", strings.TrimRight(strings.TrimLeft(m.Name, "*"), "=?"))
}

func (ModuleFunctionFieldTraits) StrictValidate(m *ModuleFunctionFieldPB) error {
	return mandatory("name", m.Name)
}

func ModuleFunctionFieldFromProto(m *ModuleFunctionFieldPB) (ModuleFunctionField, error) {
	return FromProto[ModuleFunctionField](m)
}

func StrictModuleFunctionFieldFromProto(m *ModuleFunctionFieldPB) (ModuleFunctionField, error) {
	return Strict(ModuleFunctionFieldFromProto(m))
}
