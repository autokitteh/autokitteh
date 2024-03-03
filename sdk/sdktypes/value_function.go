package sdktypes

import (
	"errors"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
	valuev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
)

type FunctionValuePB = valuev1.Function

type functionValueTraits struct{}

func (functionValueTraits) Validate(m *FunctionValuePB) error {
	return errors.Join(
		executorIDField("executor_id", m.ExecutorId),
		nameField("name", m.Name),
		objectField[ModuleFunction]("function", m.Desc),
		validateFunctionFlags(m.Flags),
	)
}

func (functionValueTraits) StrictValidate(m *FunctionValuePB) error {
	return errors.Join(
		mandatory("executor_id", m.ExecutorId),
		mandatory("name", m.Name),
	)
}

var _ objectTraits[*FunctionValuePB] = functionValueTraits{}

type FunctionValue struct {
	object[*FunctionValuePB, functionValueTraits]
}

func (FunctionValue) isConcreteValue() {}

func (f FunctionValue) ExecutorID() ExecutorID {
	return kittehs.Must1(ParseExecutorID(f.read().ExecutorId))
}

func (f FunctionValue) Data() []byte { return f.read().Data }

func (f FunctionValue) ArgNames() []string {
	fdesc := f.read().Desc
	if fdesc == nil {
		return nil
	}
	return kittehs.Transform(fdesc.Input, func(f *programv1.FunctionField) string { return f.Name })
}

func (f FunctionValue) Name() Symbol     { return kittehs.Must1(ParseSymbol(f.read().Name)) }
func (f FunctionValue) UniqueID() string { return fmt.Sprintf("%v.%s", f.ExecutorID(), f.Name()) }
func (f FunctionValue) HasFlag(flag FunctionFlag) bool {
	return kittehs.ContainedIn(f.read().Flags...)(flag.String())
}

func (v Value) IsFunction() bool           { return v.read().Function != nil }
func (v Value) GetFunction() FunctionValue { return forceFromProto[FunctionValue](v.read().Function) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsFunction() {
			return v.GetFunction()
		}
		return nil
	})
}

type FunctionFlag string

const (
	PrivilidgedFunctionFlag    FunctionFlag = "privilidged" // pass workflow context.
	PureFunctionFlag           FunctionFlag = "pure"        // do not run in an activity.
	DisablePollingFunctionFlag FunctionFlag = "no-poll"     // do not poll.
)

func (ff FunctionFlag) String() string { return string(ff) }

func validateFunctionFlags(fs []string) error {
	var errs []error

	for _, f := range fs {
		switch f {
		case PrivilidgedFunctionFlag.String(), PureFunctionFlag.String(), DisablePollingFunctionFlag.String():
			return nil
		default:
			errs = append(errs, fmt.Errorf("invalid function flag %q", f))
		}
	}

	return errors.Join(errs...)
}

func NewFunctionValue(xid ExecutorID, name string, data []byte, flags []FunctionFlag, desc ModuleFunction) (Value, error) {
	return ValueFromProto(
		&ValuePB{
			Function: &FunctionValuePB{
				ExecutorId: xid.String(),
				Name:       name,
				Desc:       desc.ToProto(),
				Data:       data,
				Flags:      kittehs.TransformToStrings(flags),
			},
		},
	)
}
