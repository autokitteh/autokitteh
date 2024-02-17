package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
)

type (
	ModulePB = programv1.Module
)

type Module = *object[*ModulePB]

var (
	ModuleFromProto       = makeFromProto(validateModule)
	StrictModuleFromProto = makeFromProto(strictValidateModule)
	ToStrictModule        = makeWithValidator(strictValidateModule)
)

func strictValidateModule(pb *programv1.Module) error {
	return validateModule(pb)
}

func validateModule(pb *programv1.Module) error {
	return nil
}

func GetModuleFunctions(m Module) map[string]ModuleFunction {
	if m == nil {
		return nil
	}

	return kittehs.Must1(kittehs.TransformMapValuesError(m.pb.Functions, StrictModuleFunctionFromProto))
}

func GetModuleVariables(m Module) map[string]ModuleVariable {
	if m == nil {
		return nil
	}

	return kittehs.Must1(kittehs.TransformMapValuesError(m.pb.Variables, StrictModuleVariableFromProto))
}

func NewModule(funcs map[string]ModuleFunction, vars map[string]ModuleVariable) Module {
	return kittehs.Must1(ModuleFromProto(&programv1.Module{
		Functions: kittehs.TransformMapValues(funcs, ToProto),
		Variables: kittehs.TransformMapValues(vars, ToProto),
	}))
}
