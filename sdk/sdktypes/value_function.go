package sdktypes

import (
	"bytes"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	modulev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/module/v1"
	valuev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
)

type FunctionValuePB = valuev1.Function

type functionValueTraits struct{ immutableObjectTrait }

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
	return kittehs.Transform(fdesc.Input, func(f *modulev1.FunctionField) string { return f.Name })
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
	PrivilegedFunctionFlag FunctionFlag = "privileged"  // pass workflow context.
	PureFunctionFlag       FunctionFlag = "pure"        // do not run in an activity.
	ConstFunctionFlag      FunctionFlag = "const"       // result is serialized in data.
	DisableAutoHeartbeat   FunctionFlag = "noheartbeat" // disable auto heartbeat.
)

func (ff FunctionFlag) String() string { return string(ff) }

func validateFunctionFlags(fs []string) error {
	var errs []error

	for _, f := range fs {
		switch f {
		case PrivilegedFunctionFlag.String(), PureFunctionFlag.String(), ConstFunctionFlag.String(), DisableAutoHeartbeat.String():
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

const (
	dataHeader  = 'D'
	errorHeader = 'E'
)

func NewConstFunctionValue(name string, data Value) (Value, error) {
	bs, err := proto.Marshal(data.ToProto())
	if err != nil {
		return InvalidValue, err
	}

	buf := bytes.NewBuffer([]byte{dataHeader})
	if _, err := buf.Write(bs); err != nil {
		return InvalidValue, err
	}

	return NewFunctionValue(
		InvalidExecutorID,
		name,
		buf.Bytes(),
		[]FunctionFlag{ConstFunctionFlag},
		InvalidModuleFunction,
	)
}

func NewConstFunctionError(name string, in error) (Value, error) {
	perr := WrapError(in)
	bs, err := proto.Marshal(perr.ToProto())
	if err != nil {
		return InvalidValue, err
	}

	buf := bytes.NewBuffer([]byte{errorHeader})
	if _, err := buf.Write(bs); err != nil {
		return InvalidValue, err
	}

	return NewFunctionValue(
		InvalidExecutorID,
		name,
		buf.Bytes(),
		[]FunctionFlag{ConstFunctionFlag},
		InvalidModuleFunction,
	)
}

func (f FunctionValue) ConstValue() (Value, error) {
	if !f.HasFlag(ConstFunctionFlag) {
		return InvalidValue, fmt.Errorf("function is not a const")
	}

	bs := f.m.Data
	if len(bs) == 0 {
		return InvalidValue, errors.New("empty data")
	}

	k, bs := bs[0], bs[1:]
	if k == dataHeader {
		var pb ValuePB
		if err := proto.Unmarshal(bs, &pb); err != nil {
			return InvalidValue, err
		}
		return ValueFromProto(&pb)
	} else if k == errorHeader {
		var pb ProgramErrorPB
		if err := proto.Unmarshal(bs, &pb); err != nil {
			return InvalidValue, err
		}

		perr, err := ProgramErrorFromProto(&pb)
		if err != nil {
			return InvalidValue, err
		}

		return InvalidValue, perr.ToError()
	}

	return InvalidValue, fmt.Errorf("invalid data type %q", k)
}
