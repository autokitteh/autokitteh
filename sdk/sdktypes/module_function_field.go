package sdktypes

import (
	"strings"

	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
)

type ModuleFunctionField struct {
	object[*ModuleFunctionFieldPB, ModuleFunctionFieldTraits]
}

type ModuleFunctionFieldPB = programv1.FunctionField

type ModuleFunctionFieldTraits struct{}

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
