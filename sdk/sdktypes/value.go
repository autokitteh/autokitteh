package sdktypes

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/araddon/dateparse"

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
		objectField[FunctionValue]("function", m.Function),
		objectField[NothingValue]("nothing", m.Nothing),
		objectField[IntegerValue]("integer", m.Integer),
		objectField[FloatValue]("float", m.Float),
		objectField[BooleanValue]("boolean", m.Boolean),
		objectField[BytesValue]("bytes", m.Bytes),
		objectField[DurationValue]("duration", m.Duration),
		objectField[TimeValue]("time", m.Time),
		objectField[CustomValue]("custom", m.Custom),
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
	case CustomValue:
		return kittehs.Must1(NewCustomValue(cv.ExecutorID(), cv.Data(), cv.Value()))
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

func (v Value) ToDuration() (time.Duration, error) {
	switch v := v.Concrete().(type) {
	case CustomValue:
		return v.Value().ToDuration()
	case DurationValue:
		return v.Value(), nil
	case IntegerValue:
		return time.Second * time.Duration(v.Value()), nil
	case FloatValue:
		return time.Duration(float64(time.Second) * v.Value()), nil
	case StringValue:
		return time.ParseDuration(v.Value())
	default:
		return 0, errors.New("value not convertible to duration")
	}
}

func (v Value) ToTime() (time.Time, error) {
	switch v := v.Concrete().(type) {
	case CustomValue:
		return v.Value().ToTime()
	case TimeValue:
		return v.Value(), nil
	case StringValue:
		return dateparse.ParseAny(v.Value())
	case IntegerValue:
		return time.Unix(v.Value(), 0), nil
	default:
		return time.Time{}, errors.New("value not convertible to time")
	}
}

func (v Value) ToString() (string, error) {
	switch v := v.Concrete().(type) {
	case CustomValue:
		return v.Value().ToString()
	case StringValue:
		return v.Value(), nil
	case IntegerValue:
		return fmt.Sprintf("%d", v.Value()), nil
	case FloatValue:
		return fmt.Sprintf("%f", v.Value()), nil
	case BooleanValue:
		return fmt.Sprintf("%t", v.Value()), nil
	case DurationValue:
		return v.Value().String(), nil
	case TimeValue:
		return v.Value().String(), nil
	case BytesValue:
		return base64.StdEncoding.EncodeToString(v.Value()), nil
	case SymbolValue:
		return v.Symbol().String(), nil
	default:
		return "", errors.New("not convertible to string")
	}
}

func (v Value) ToStringValuesMap() (map[string]Value, error) {
	switch v := v.Concrete().(type) {
	case CustomValue:
		return v.Value().ToStringValuesMap()
	case DictValue:
		return v.ToStringValuesMap()
	case StructValue:
		return v.Fields(), nil
	case ModuleValue:
		return v.Members(), nil
	default:
		return nil, errors.New("not convertible to map")
	}
}

func (v Value) Unwrap() (any, error)     { return UnwrapValue(v) }
func (v Value) UnwrapInto(dst any) error { return UnwrapValueInto(dst, v) }

// An unwrapper that is alway safe to serialize to string afterwards.
var valueStringUnwrapper = ValueWrapper{
	SafeForJSON: true,
	Preunwrap: func(v Value) (Value, error) {
		if v.IsFunction() {
			return NewStringValuef("|function: %v|", v.GetFunction().Name()), nil
		}

		return v, nil
	},
}

func ValueProtoToJSONStringValue(pb *ValuePB) (*ValuePB, error) {
	if pb == nil {
		return nil, nil
	}

	v, err := ValueFromProto(pb)
	if err != nil {
		return nil, fmt.Errorf("decode proto: %w", err)
	}

	u, err := valueStringUnwrapper.Unwrap(v)
	if err != nil {
		return nil, fmt.Errorf("unwrap value: %w", err)
	}

	j, err := json.Marshal(u)
	if err != nil {
		return nil, fmt.Errorf("marshal JSON: %w", err)
	}

	return NewStringValue(string(j)).ToProto(), nil
}
