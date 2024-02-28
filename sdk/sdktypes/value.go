package sdktypes

import (
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	"google.golang.org/protobuf/proto"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
	valuesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type ValuePB = valuesv1.Value

type Value = *object[*ValuePB]

var (
	MustValueFromProto = makeMustFromProto(ValidateValuePB)
	ToStrictValue      = makeWithValidator(StrictValidateValuePB)
)

func ValueFromProto(pb *ValuePB) (Value, error) {
	if pb == nil {
		return nil, nil
	}
	return fromProto(pb, ValidateValuePB)
}

func MustNewValue(v Object) Value { return kittehs.Must1(NewValue(v)) }

func NewValue(v Object) (Value, error) {
	pb, err := newValuePB(v)
	if err != nil {
		return nil, err
	}

	return ValueFromProto(pb)
}

func StrictValueFromProto(pb *ValuePB) (Value, error) {
	return fromProto(pb, StrictValidateValuePB)
}

func StrictValidateValuePB(pb *ValuePB) error {
	if pb.Type == nil {
		return fmt.Errorf("%w: empty value", sdkerrors.ErrInvalidArgument)
	}

	return ValidateValuePB(pb)
}

func ValidateValuePB(pb *ValuePB) error {
	if _, err := getValue(pb); err != nil {
		return err
	}

	return nil
}

func GetValue(v Value) Object { return kittehs.Must1(getValue(v.pb)) }

func StringValueMapToProto(m map[string]Value) map[string]*ValuePB {
	return kittehs.TransformMap(m, func(k string, v Value) (string, *ValuePB) {
		return k, v.ToProto()
	})
}

func isValueOfType[T Object](v Value) bool {
	if v == nil {
		return false
	}

	_, ok := GetValue(v).(T)
	return ok
}

var (
	IsNothingValue  = isValueOfType[NothingValue]
	IsBooleanValue  = isValueOfType[BooleanValue]
	IsStringValue   = isValueOfType[StringValue]
	IsIntegerValue  = isValueOfType[IntegerValue]
	IsBytesValue    = isValueOfType[BytesValue]
	IsFloatValue    = isValueOfType[FloatValue]
	IsDurationValue = isValueOfType[DurationValue]
	IsTimeValue     = isValueOfType[TimeValue]
	IsSymbolValue   = isValueOfType[SymbolValue]
	IsListValue     = isValueOfType[ListValue]
	IsSetValue      = isValueOfType[SetValue]
	IsDictValue     = isValueOfType[DictValue]
	IsStructValue   = isValueOfType[StructValue]
	IsModuleValue   = isValueOfType[ModuleValue]
	IsFunctionValue = isValueOfType[FunctionValue]
)

func GetBooleanValue(v Value) bool           { return GetValue(v).(BooleanValue).pb.V }
func GetStringValue(v Value) string          { return GetValue(v).(StringValue).pb.V }
func GetIntegerValue(v Value) int64          { return GetValue(v).(IntegerValue).pb.V }
func GetBytesValue(v Value) []byte           { return GetValue(v).(BytesValue).pb.V }
func GetFloatValue(v Value) float64          { return GetValue(v).(FloatValue).pb.V }
func GetDurationValue(v Value) time.Duration { return GetValue(v).(DurationValue).pb.V.AsDuration() }
func GetTimeValue(v Value) time.Time         { return GetValue(v).(TimeValue).pb.V.AsTime() }

func GetSymbolValue(v Value) Symbol {
	return kittehs.Must1(StrictParseSymbol(GetValue(v).(SymbolValue).pb.Name))
}

func GetListValue(v Value) []Value {
	if !IsListValue(v) {
		return nil
	}

	return kittehs.Must1(kittehs.TransformError(GetValue(v).(ListValue).pb.Vs, StrictValueFromProto))
}

func GetSetValue(v Value) []Value {
	return kittehs.Must1(kittehs.TransformError(GetValue(v).(SetValue).pb.Vs, StrictValueFromProto))
}

func GetDictValue(v Value) []*DictValueItem {
	if v == nil {
		return nil
	}

	if !IsDictValue(v) {
		return nil
	}

	return kittehs.Transform(
		GetValue(v).(DictValue).pb.Items,
		func(pb *valuesv1.Dict_Item) *DictValueItem {
			return &DictValueItem{
				K: kittehs.Must1(StrictValueFromProto(pb.K)),
				V: kittehs.Must1(StrictValueFromProto(pb.V)),
			}
		},
	)
}

func GetDictValueKeys(v Value) []Value {
	if v == nil {
		return nil
	}

	return kittehs.Transform(
		GetValue(v).(DictValue).pb.Items,
		func(pb *valuesv1.Dict_Item) Value {
			return kittehs.Must1(StrictValueFromProto(pb.K))
		},
	)
}

func GetDictValueLen(v Value) int {
	if v == nil {
		return 0
	}

	if !IsDictValue(v) {
		return 0
	}

	return len(GetValue(v).(DictValue).pb.Items)
}

func DictValueToStringsMap(v Value) (map[string]Value, error) {
	if v == nil {
		return nil, nil
	}

	return kittehs.ListToMapError(
		GetDictValue(v),
		func(i *DictValueItem) (string, Value, error) {
			if !IsStringValue(i.K) {
				return "", nil, fmt.Errorf("key is not a string: %w", sdkerrors.ErrInvalidArgument)
			}
			return GetStringValue(i.K), i.V, nil
		},
	)
}

func GetModuleValue(v Value) (Symbol, map[string]Value) {
	mv := GetValue(v).(ModuleValue).pb
	return kittehs.Must1(StrictParseSymbol(mv.Name)), kittehs.Must1(kittehs.TransformMapValuesError(mv.Members, StrictValueFromProto))
}

func GetStructValue(v Value) (Value, map[string]Value) {
	mv := GetValue(v).(StructValue).pb
	return kittehs.Must1(StrictValueFromProto(mv.Ctor)),
		kittehs.Must1(kittehs.TransformMapValuesError(mv.Fields, StrictValueFromProto))
}

func GetFunctionValue(v Value) FunctionValue {
	if v == nil {
		return nil
	}
	return kittehs.Must1(FunctionValueFromProto(GetValue(v).(FunctionValue).pb))
}

func GetFunctionValueData(v Value) []byte { return GetFunctionValue(v).pb.Data }

func GetFunctionValueExecutorID(v Value) ExecutorID {
	return kittehs.Must1(ParseExecutorID(GetFunctionValue(v).pb.ExecutorId))
}

func GetFunctionValueName(v Value) Symbol {
	return kittehs.Must1(ParseSymbol(GetFunctionValue(v).pb.Name))
}

func GetFunctionValueFlags(v Value) []FunctionFlag {
	return kittehs.Transform(GetFunctionValue(v).pb.Flags, func(s string) FunctionFlag { return FunctionFlag(s) })
}

func GetFunctionValueArgsNames(v Value) []string {
	fdesc := GetFunctionValue(v).pb.Desc
	if fdesc == nil {
		return nil
	}
	return kittehs.Transform(fdesc.Input, func(f *programv1.FunctionField) string { return f.Name })
}

func FunctionValueHasFlag(v Value, flag FunctionFlag) bool {
	return kittehs.ContainedIn(GetFunctionValueFlags(v)...)(flag)
}

func GetFunctionValueUniqueID(v Value) string {
	return fmt.Sprintf("%v.%s", GetFunctionValueExecutorID(v), GetFunctionValueName(v))
}

func FunctionValueHasExecutorID(v Value) bool { return GetFunctionValue(v).pb.ExecutorId != "" }

func EqualValues(a, b Value) bool { return proto.Equal(a.pb, b.pb) }

func GetValueLength(v Value) (int, error) {
	switch pb := v.pb.Type.(type) {
	case *valuesv1.Value_Dict:
		return len(pb.Dict.Items), nil
	case *valuesv1.Value_Bytes:
		return len(pb.Bytes.V), nil
	case *valuesv1.Value_String_:
		return len(pb.String_.V), nil
	case *valuesv1.Value_List:
		return len(pb.List.Vs), nil
	case *valuesv1.Value_Set:
		return len(pb.Set.Vs), nil
	default:
		return 0, sdkerrors.ErrInvalidArgument
	}
}

func ValueToStringValuesMap(v Value) (map[string]Value, error) {
	if v == nil {
		return nil, nil
	}

	switch GetValue(v).(type) {
	case DictValue:
		return DictValueToStringMap(v)
	case StructValue:
		return StructValueToStringMap(v)
	default:
		return nil, sdkerrors.ErrInvalidArgument
	}
}

func ValueToDuration(v Value) (time.Duration, error) {
	switch v := GetValue(v).(type) {
	case DurationValue:
		return v.pb.V.AsDuration(), nil
	case IntegerValue:
		return time.Second * time.Duration(v.pb.V), nil
	case FloatValue:
		return time.Duration(float64(time.Second) * v.pb.V), nil
	case StringValue:
		return time.ParseDuration(v.pb.V)
	default:
		return 0, fmt.Errorf("value not convertible to duration")
	}
}

func ValueToTime(v Value) (time.Time, error) {
	switch v := GetValue(v).(type) {
	case TimeValue:
		return v.pb.V.AsTime(), nil
	case StringValue:
		return dateparse.ParseAny(v.pb.V)
	case IntegerValue:
		return time.Unix(v.pb.V, 0), nil
	default:
		return time.Time{}, fmt.Errorf("value not convertible to time")
	}
}
