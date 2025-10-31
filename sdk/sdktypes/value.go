package sdktypes

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/araddon/dateparse"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	valuev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
)

type Value struct{ object[*ValuePB, ValueTraits] }

func init() { registerObject[Value]() }

var InvalidValue Value

type ValuePB = valuev1.Value

type ValueTraits struct{ immutableObjectTrait }

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
		objectField[BigIntegerValue]("big_integer", m.BigInteger),
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
	case BigIntegerValue:
		return NewBigIntegerValue(cv.Value())
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

func errCannotConvert(v Value, to string) error {
	return sdkerrors.NewInvalidArgumentError("cannot convert %s to %s", v.Type(), to)
}

func (v Value) Type() string {
	if !v.IsValid() {
		return "invalid"
	}

	switch v.Concrete().(type) {
	case NothingValue:
		return "nothing"
	case IntegerValue:
		return "integer"
	case FloatValue:
		return "float"
	case BooleanValue:
		return "boolean"
	case StringValue:
		return "string"
	case BytesValue:
		return "bytes"
	case DurationValue:
		return "duration"
	case TimeValue:
		return "time"
	case SymbolValue:
		return "symbol"
	case ListValue:
		return "list"
	case SetValue:
		return "set"
	case DictValue:
		return "dict"
	case StructValue:
		return "struct"
	case ModuleValue:
		return "module"
	case CustomValue:
		return "custom"
	case BigIntegerValue:
		return "big_integer"
	default:
		return "unknown"
	}
}

func (v Value) ToInt64() (int64, error) {
	switch vv := v.Concrete().(type) {
	case CustomValue:
		return vv.Value().ToInt64()
	case DurationValue:
		return int64(vv.Value()), nil
	case FloatValue:
		return int64(vv.Value()), nil
	case IntegerValue:
		return vv.Value(), nil
	case BigIntegerValue:
		bi := vv.Value()
		if !bi.IsInt64() {
			return 0, sdkerrors.NewInvalidArgumentError("big integer value %s overflows int64", bi.String())
		}
		return bi.Int64(), nil
	default:
		return 0, errCannotConvert(v, "int64")
	}
}

func (v Value) ToBigInteger() (*big.Int, error) {
	switch vv := v.Concrete().(type) {
	case CustomValue:
		return vv.Value().ToBigInteger()
	case DurationValue:
		return big.NewInt(int64(vv.Value())), nil
	case FloatValue:
		return big.NewInt(int64(vv.Value())), nil
	case IntegerValue:
		return big.NewInt(vv.Value()), nil
	case BigIntegerValue:
		return vv.Value(), nil
	default:
		return nil, errCannotConvert(v, "big.Int")
	}
}

func (v Value) ToFloat64() (float64, error) {
	switch vv := v.Concrete().(type) {
	case CustomValue:
		return vv.Value().ToFloat64()
	case DurationValue:
		return float64(vv.Value()), nil
	case IntegerValue:
		return float64(vv.Value()), nil
	case FloatValue:
		return vv.Value(), nil
	case BigIntegerValue:
		bi := vv.Value()
		f, _ := bi.Float64()
		if math.IsInf(f, 0) || math.IsNaN(f) {
			return 0, sdkerrors.NewInvalidArgumentError("big integer value %s overflows float64", bi.String())
		}
		return f, nil
	default:
		return 0, errCannotConvert(v, "float64")
	}
}

func (v Value) ToDuration() (time.Duration, error) {
	switch vv := v.Concrete().(type) {
	case CustomValue:
		return vv.Value().ToDuration()
	case DurationValue:
		return vv.Value(), nil
	case IntegerValue:
		return time.Second * time.Duration(vv.Value()), nil
	case FloatValue:
		return time.Duration(float64(time.Second) * vv.Value()), nil
	case BigIntegerValue:
		bi := vv.Value()
		if !bi.IsInt64() {
			return 0, sdkerrors.NewInvalidArgumentError("big integer value %s overflows duration", bi.String())
		}
		return time.Duration(bi.Int64()) * time.Second, nil
	case StringValue:
		return time.ParseDuration(vv.Value())
	default:
		return 0, errCannotConvert(v, "Duration")
	}
}

func (v Value) ToTime() (time.Time, error) {
	switch vv := v.Concrete().(type) {
	case CustomValue:
		return vv.Value().ToTime()
	case TimeValue:
		return vv.Value(), nil
	case StringValue:
		return dateparse.ParseAny(vv.Value())
	case IntegerValue:
		return time.Unix(vv.Value(), 0), nil
	case BigIntegerValue:
		bi := vv.Value()
		if !bi.IsInt64() {
			return time.Time{}, sdkerrors.NewInvalidArgumentError("big integer value %s overflows time", bi.String())
		}
		return time.Unix(bi.Int64(), 0), nil
	case FloatValue:
		sec := int64(vv.Value())
		nsec := int64((vv.Value() - float64(sec)) * float64(time.Second))
		return time.Unix(sec, nsec), nil
	default:
		return time.Time{}, errCannotConvert(v, "Time")
	}
}

func (v Value) ToString() (string, error) {
	switch vv := v.Concrete().(type) {
	case CustomValue:
		return vv.Value().ToString()
	case StringValue:
		return vv.Value(), nil
	case IntegerValue:
		return strconv.FormatInt(vv.Value(), 10), nil
	case BigIntegerValue:
		return vv.Value().String(), nil
	case FloatValue:
		return fmt.Sprintf("%f", vv.Value()), nil
	case BooleanValue:
		return strconv.FormatBool(vv.Value()), nil
	case DurationValue:
		return vv.Value().String(), nil
	case TimeValue:
		return vv.Value().String(), nil
	case BytesValue:
		return base64.StdEncoding.EncodeToString(vv.Value()), nil
	case SymbolValue:
		return vv.Symbol().String(), nil
	default:
		return "", errCannotConvert(v, "String")
	}
}

func (v Value) ToStringValuesMap() (map[string]Value, error) {
	switch vv := v.Concrete().(type) {
	case CustomValue:
		return vv.Value().ToStringValuesMap()
	case DictValue:
		return vv.ToStringValuesMap()
	case StructValue:
		return vv.Fields(), nil
	case ModuleValue:
		return vv.Members(), nil
	default:
		return nil, errCannotConvert(v, "map")
	}
}

func (v Value) Unwrap() (any, error)     { return UnwrapValue(v) }
func (v Value) UnwrapInto(dst any) error { return UnwrapValueInto(dst, v) }

// An unwrapper that is always safe to serialize to string afterwards.
var valueStringUnwrapper = ValueWrapper{
	SafeForJSON:         true,
	UnwrapStructsAsJSON: true,
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
