package sdktypes

import (
	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
)

type (
	ModuleVariablePB = programv1.Variable
)

type ModuleVariable = *object[*ModuleVariablePB]

var (
	ModuleVariableFromProto       = makeFromProto(validateModuleVariable)
	StrictModuleVariableFromProto = makeFromProto(strictValidateModuleVariable)
	ToStrictModuleVariable        = makeWithValidator(strictValidateModuleVariable)
)

func strictValidateModuleVariable(pb *programv1.Variable) error {
	return validateModuleVariable(pb)
}

func validateModuleVariable(pb *programv1.Variable) error {
	return nil
}

func GetModuleVariableDescription(i ModuleVariable) string {
	if i == nil {
		return ""
	}

	return i.pb.Description
}
