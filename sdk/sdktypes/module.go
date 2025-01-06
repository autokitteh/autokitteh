package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	modulev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/module/v1"
)

type Module struct {
	object[*ModulePB, ModuleTraits]
}

type ModulePB = modulev1.Module

type ModuleTraits struct{ immutableObjectTrait }

func (ModuleTraits) Validate(m *ModulePB) error {
	return errors.Join(
		objectsMapField[ModuleFunction]("functions", m.Functions),
		objectsMapField[ModuleVariable]("variables", m.Variables),
	)
}

func (ModuleTraits) StrictValidate(m *ModulePB) error { return nil }

var InvalidModule Module

func ModuleFromProto(m *ModulePB) (Module, error)       { return FromProto[Module](m) }
func StrictModuleFromProto(m *ModulePB) (Module, error) { return Strict(ModuleFromProto(m)) }

func NewModule(fs map[string]ModuleFunction, vs map[string]ModuleVariable) (Module, error) {
	return FromProto[Module](&ModulePB{
		Functions: kittehs.TransformMapValues(fs, ToProto),
		Variables: kittehs.TransformMapValues(vs, ToProto),
	})
}
