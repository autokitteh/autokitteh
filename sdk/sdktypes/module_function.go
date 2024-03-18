package sdktypes

import (
	"errors"

	modulev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/module/v1"
)

type ModuleFunction struct {
	object[*ModuleFunctionPB, ModuleFunctionTraits]
}

type ModuleFunctionPB = modulev1.Function

type ModuleFunctionTraits struct{}

var InvalidModuleFunction ModuleFunction

func (ModuleFunctionTraits) Validate(m *ModuleFunctionPB) error {
	return errors.Join(
		urlField("url", m.DocumentationUrl),
		objectsSliceField[ModuleFunctionField]("input", m.Input),
		objectsSliceField[ModuleFunctionField]("output", m.Output),
	)
}

func (ModuleFunctionTraits) StrictValidate(m *ModuleFunctionPB) error {
	return nil
}

func ModuleFunctionFromProto(m *ModuleFunctionPB) (ModuleFunction, error) {
	return FromProto[ModuleFunction](m)
}

func StrictModuleFunctionFromProto(m *ModuleFunctionPB) (ModuleFunction, error) {
	return Strict(ModuleFunctionFromProto(m))
}
