package sdktypes

import (
	"encoding/base64"
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

func (v Value) ToDuration() (time.Duration, error) {
	switch v := v.Concrete().(type) {
	case DurationValue:
		return v.Value(), nil
	case IntegerValue:
		return time.Second * time.Duration(v.Value()), nil
	case FloatValue:
		return time.Duration(float64(time.Second) * v.Value()), nil
	case StringValue:
		return time.ParseDuration(v.Value())
	default:
		return 0, fmt.Errorf("value not convertible to duration")
	}
}

func (v Value) ToTime() (time.Time, error) {
	switch v := v.Concrete().(type) {
	case TimeValue:
		return v.Value(), nil
	case StringValue:
		return dateparse.ParseAny(v.Value())
	case IntegerValue:
		return time.Unix(v.Value(), 0), nil
	default:
		return time.Time{}, fmt.Errorf("value not convertible to time")
	}
}

func (v Value) ToString() (string, error) {
	switch v := v.Concrete().(type) {
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
