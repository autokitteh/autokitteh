package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	valuev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
)

type Value struct{ object[*ValuePB, ValueTraits] }

var InvalidValue Value

type ValuePB = valuev1.Value

type ValueTraits struct{}

func validateValue(m *ValuePB) error {
	return errors.Join(
		objectField[SymbolValue]("symbol", m.Symbol),
		objectField[ListValue]("list", m.List),
		objectField[SetValue]("set", m.Set),
		objectField[DictValue]("dict", m.Dict),
		objectField[StructValue]("struct", m.Struct),
		objectField[ModuleValue]("module", m.Module),
	)
}

func (ValueTraits) Validate(m *ValuePB) error       { return validateValue(m) }
func (ValueTraits) StrictValidate(m *ValuePB) error { return oneOfMessage(m) }

func ValueFromProto(m *ValuePB) (Value, error)       { return FromProto[Value](m) }
func StrictValueFromProto(m *ValuePB) (Value, error) { return Strict(ValueFromProto(m)) }

func NewValue(cv concreteValue) Value {
	switch cv := cv.(type) {
	case NothingValue:
		return Nothing
	case IntegerValue:
		return NewIntegerValue(cv.Value())
	case FloatValue:
		return NewFloatValue(cv.Value())
	case BooleanValue:
		return NewBooleanValue(cv.Value())
	case StringValue:
		return NewStringValue(cv.Value())
	case BytesValue:
		return NewBytesValue(cv.Value())
	case DurationValue:
		return NewDurationValue(cv.Value())
	case TimeValue:
		return NewTimeValue(cv.Value())
	case SymbolValue:
		return NewSymbolValue(cv.Symbol())
	case ListValue:
		return kittehs.Must1(NewListValue(cv.Values()))
	case SetValue:
		return kittehs.Must1(NewSetValue(cv.Values()))
	case DictValue:
		return kittehs.Must1(NewDictValue(cv.Items()))
	case StructValue:
		return kittehs.Must1(NewStructValue(cv.Ctor(), cv.Fields()))
	case ModuleValue:
		return kittehs.Must1(NewModuleValue(cv.Name(), cv.Members()))
	default:
		sdklogger.DPanic("unknown concrete value type")
	}

	return Value{}
}

var valueGetters = []func(Value) concreteValue{}

func registerValueGetter(f func(Value) concreteValue) {
	valueGetters = append(valueGetters, f)
}

func (v Value) Concrete() concreteValue {
	if !v.IsValid() {
		return nil
	}

	for _, get := range valueGetters {
		if c := get(v); c != nil {
			return c
		}
	}

	sdklogger.DPanic("unknown value type")

	return nil
}
