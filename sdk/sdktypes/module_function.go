package sdktypes

import (
	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
)

type (
	ModuleFunctionPB      = programv1.Function
	ModuleFunctionFieldPB = programv1.FunctionField
)

type ModuleFunction = *object[*ModuleFunctionPB]

var (
	ModuleFunctionFromProto       = makeFromProto(validateModuleFunction)
	StrictModuleFunctionFromProto = makeFromProto(strictValidateModuleFunction)
	ToStrictModuleFunction        = makeWithValidator(strictValidateModuleFunction)
)

func strictValidateModuleFunction(pb *programv1.Function) error {
	return validateModuleFunction(pb)
}

func validateModuleFunction(pb *programv1.Function) error {
	return nil
}

func GetModuleFunctionDescription(i ModuleFunction) string {
	if i == nil {
		return ""
	}

	return i.pb.Description
}
