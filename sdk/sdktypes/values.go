package sdktypes

import (
	"fmt"
	"sort"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	valuesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func newValuePB(v Object) (*ValuePB, error) {
	switch pb := v.toMessage().(type) {
	case *valuesv1.Nothing:
		return &ValuePB{Type: &valuesv1.Value_Nothing{}}, nil
	case *valuesv1.Boolean:
		return &ValuePB{Type: &valuesv1.Value_Boolean{Boolean: pb}}, nil
	case *valuesv1.String:
		return &ValuePB{Type: &valuesv1.Value_String_{String_: pb}}, nil
	case *valuesv1.Integer:
		return &ValuePB{Type: &valuesv1.Value_Integer{Integer: pb}}, nil
	case *valuesv1.Bytes:
		return &ValuePB{Type: &valuesv1.Value_Bytes{Bytes: pb}}, nil
	case *valuesv1.Float:
		return &ValuePB{Type: &valuesv1.Value_Float{Float: pb}}, nil
	case *valuesv1.Symbol:
		return &ValuePB{Type: &valuesv1.Value_Symbol{Symbol: pb}}, nil
	case *valuesv1.Duration:
		return &ValuePB{Type: &valuesv1.Value_Duration{Duration: pb}}, nil
	case *valuesv1.Time:
		return &ValuePB{Type: &valuesv1.Value_Time{Time: pb}}, nil
	case *valuesv1.List:
		return &ValuePB{Type: &valuesv1.Value_List{List: pb}}, nil
	case *valuesv1.Set:
		return &ValuePB{Type: &valuesv1.Value_Set{Set: pb}}, nil
	case *valuesv1.Dict:
		return &ValuePB{Type: &valuesv1.Value_Dict{Dict: pb}}, nil
	case *valuesv1.Module:
		return &ValuePB{Type: &valuesv1.Value_Module{Module: pb}}, nil
	case *valuesv1.Struct:
		return &ValuePB{Type: &valuesv1.Value_Struct{Struct: pb}}, nil
	case *valuesv1.Function:
		return &ValuePB{Type: &valuesv1.Value_Function{Function: pb}}, nil
	default:
		return nil, fmt.Errorf("unrecognized type: %w", sdkerrors.ErrInvalidArgument)
	}
}

func getValue(pb *ValuePB) (Object, error) {
	if pb == nil {
		return nil, nil
	}

	switch pb := pb.Type.(type) {
	case *valuesv1.Value_Nothing:
		return nothingValue, nil
	case *valuesv1.Value_Boolean:
		return BooleanValueFromProto(pb.Boolean)
	case *valuesv1.Value_String_:
		return StringValueFromProto(pb.String_)
	case *valuesv1.Value_Integer:
		return IntegerValueFromProto(pb.Integer)
	case *valuesv1.Value_Bytes:
		return BytesValueFromProto(pb.Bytes)
	case *valuesv1.Value_Float:
		return FloatValueFromProto(pb.Float)
	case *valuesv1.Value_Symbol:
		return SymbolValueFromProto(pb.Symbol)
	case *valuesv1.Value_Duration:
		return DurationValueFromProto(pb.Duration)
	case *valuesv1.Value_Time:
		return TimeValueFromProto(pb.Time)
	case *valuesv1.Value_List:
		return ListValueFromProto(pb.List)
	case *valuesv1.Value_Set:
		return SetValueFromProto(pb.Set)
	case *valuesv1.Value_Dict:
		return DictValueFromProto(pb.Dict)
	case *valuesv1.Value_Module:
		return ModuleValueFromProto(pb.Module)
	case *valuesv1.Value_Struct:
		return StructValueFromProto(pb.Struct)
	case *valuesv1.Value_Function:
		return FunctionValueFromProto(pb.Function)
	default:
		return nil, fmt.Errorf("%w: unrecognized type: %T", sdkerrors.ErrInvalidArgument, pb)
	}
}

// Values

type (
	NothingValue  = *object[*valuesv1.Nothing]
	BooleanValue  = *object[*valuesv1.Boolean]
	StringValue   = *object[*valuesv1.String]
	IntegerValue  = *object[*valuesv1.Integer]
	BytesValue    = *object[*valuesv1.Bytes]
	FloatValue    = *object[*valuesv1.Float]
	SymbolValue   = *object[*valuesv1.Symbol]
	DurationValue = *object[*valuesv1.Duration]
	TimeValue     = *object[*valuesv1.Time]
	ListValue     = *object[*valuesv1.List]
	SetValue      = *object[*valuesv1.Set]
	DictValue     = *object[*valuesv1.Dict]
	ModuleValue   = *object[*valuesv1.Module]
	StructValue   = *object[*valuesv1.Struct]
	FunctionValue = *object[*valuesv1.Function]
)

var (
	nothingValue    = makeMustFromProto[*valuesv1.Nothing](nil)(&valuesv1.Nothing{})
	NewNothingValue = func() Value { return MustNewValue(nothingValue) }

	BooleanValueFromProto = makeFromProto[*valuesv1.Boolean](nil)
	NewBooleanValue       = func(v bool) Value { return MustNewValue(kittehs.Must1(BooleanValueFromProto(&valuesv1.Boolean{V: v}))) }

	StringValueFromProto = makeFromProto[*valuesv1.String](nil)
	NewStringValue       = func(v string) Value { return MustNewValue(kittehs.Must1(StringValueFromProto(&valuesv1.String{V: v}))) }

	IntegerValueFromProto = makeFromProto[*valuesv1.Integer](nil)
	NewIntegerValue       = func(v int64) Value {
		return MustNewValue(kittehs.Must1(IntegerValueFromProto(&valuesv1.Integer{V: v})))
	}

	BytesValueFromProto = makeFromProto[*valuesv1.Bytes](nil)
	NewBytesValue       = func(v []byte) Value { return MustNewValue(kittehs.Must1(BytesValueFromProto(&valuesv1.Bytes{V: v}))) }

	FloatValueFromProto = makeFromProto[*valuesv1.Float](nil)
	NewFloatValue       = func(v float64) Value { return MustNewValue(kittehs.Must1(FloatValueFromProto(&valuesv1.Float{V: v}))) }

	SymbolValueFromProto = makeFromProto[*valuesv1.Symbol](nil)
	NewSymbolValue       = func(v Symbol) Value {
		return MustNewValue(kittehs.Must1(SymbolValueFromProto(&valuesv1.Symbol{Name: v.String()})))
	}

	DurationValueFromProto = makeFromProto[*valuesv1.Duration](nil)
	NewDurationValue       = func(v time.Duration) Value {
		return MustNewValue(kittehs.Must1(DurationValueFromProto(&valuesv1.Duration{V: durationpb.New(v)})))
	}

	TimeValueFromProto = makeFromProto[*valuesv1.Time](nil)
	NewTimeValue       = func(v time.Time) Value {
		return MustNewValue(kittehs.Must1(TimeValueFromProto(&valuesv1.Time{V: timestamppb.New(v)})))
	}
)

func ListValueFromProto(pb *valuesv1.List) (ListValue, error) {
	return makeFromProto(func(pb *valuesv1.List) error {
		_, err := kittehs.ValidateList(pb.Vs, StrictValidateValuePB)
		return err
	})(pb)
}

func NewListValue(vs []Value) Value {
	return MustNewValue(kittehs.Must1(ListValueFromProto(&valuesv1.List{
		Vs: kittehs.Transform(vs, func(v Value) *ValuePB { return v.ToProto() }),
	})))
}

func SetValueFromProto(pb *valuesv1.Set) (SetValue, error) {
	return makeFromProto(func(pb *valuesv1.Set) error {
		_, err := kittehs.ValidateList(pb.Vs, StrictValidateValuePB)
		return err
	})(pb)
}

func NewSetValue(vs []Value) Value {
	hashes := make(map[string]bool)
	return MustNewValue(kittehs.Must1(SetValueFromProto(&valuesv1.Set{
		Vs: kittehs.FilterNils(
			kittehs.Transform(
				vs,
				func(v Value) *ValuePB {
					// TODO: This must be terribly slow. Find a better way.
					hash := GetObjectHash(v)
					if hashes[hash] {
						return nil
					}
					hashes[hash] = true
					return v.ToProto()
				},
			),
		),
	})))
}

func DictValueFromProto(pb *valuesv1.Dict) (DictValue, error) {
	// TODO: validate key uniqueness?

	return makeFromProto(func(pb *valuesv1.Dict) error {
		return kittehs.FirstError(kittehs.Transform(pb.Items, func(pb *valuesv1.Dict_Item) error {
			if _, err := StrictValueFromProto(pb.K); err != nil {
				return fmt.Errorf("key: %w", err)
			}

			if _, err := StrictValueFromProto(pb.V); err != nil {
				return fmt.Errorf("value: %w", err)
			}

			return nil
		}))
	})(pb)
}

type DictValueItem struct{ K, V Value }

// Creates a new Dict Value. Dictionary items are stored sorted by their string representation
// of their key.
func NewDictValue(items []*DictValueItem) Value {
	pbitems := kittehs.Transform(items, func(kv *DictValueItem) *valuesv1.Dict_Item {
		return &valuesv1.Dict_Item{
			K: kv.K.ToProto(),
			V: kv.V.ToProto(),
		}
	})

	sort.SliceStable(pbitems, func(i, j int) bool { return pbitems[i].K.String() < pbitems[j].K.String() })

	return MustNewValue(kittehs.Must1(DictValueFromProto(&valuesv1.Dict{
		Items: pbitems,
	})))
}

func DictValueToStringMap(v Value) (map[string]Value, error) {
	vv := GetDictValue(v)
	if vv == nil {
		return nil, nil
	}

	m := make(map[string]Value, len(vv))
	for _, i := range vv {
		if !IsStringValue(i.K) {
			return nil, sdkerrors.ErrInvalidArgument
		}

		m[GetStringValue(i.K)] = i.V
	}

	return m, nil
}

func NewDictValueFromStringMap(kvs map[string]Value) Value {
	items := make([]*DictValueItem, 0, len(kvs))
	for k, v := range kvs {
		items = append(items, &DictValueItem{K: NewStringValue(k), V: v})
	}
	return NewDictValue(items)
}

func ModuleValueFromProto(pb *valuesv1.Module) (ModuleValue, error) {
	return makeFromProto(func(pb *valuesv1.Module) error {
		if _, err := StrictParseSymbol(pb.Name); err != nil {
			return fmt.Errorf("name: %w", err)
		}

		for k, v := range pb.Members {
			if _, err := StrictParseSymbol(k); err != nil {
				return fmt.Errorf("member key: %w", err)
			}

			if _, err := StrictValueFromProto(v); err != nil {
				return fmt.Errorf("member value: %w", err)
			}
		}

		return nil
	})(pb)
}

func NewModuleValue(name Symbol, members map[string]Value) Value {
	return kittehs.Must1(NewValue(kittehs.Must1(ModuleValueFromProto(&valuesv1.Module{
		Name: name.String(),
		Members: kittehs.TransformMapValues(
			members,
			func(v Value) *valuesv1.Value { return v.ToProto() },
		),
	}))))
}

func StructValueFromProto(pb *valuesv1.Struct) (StructValue, error) {
	return makeFromProto(func(pb *valuesv1.Struct) error {
		if _, err := StrictValueFromProto(pb.Ctor); err != nil {
			return fmt.Errorf("name: %w", err)
		}

		for k, v := range pb.Fields {
			if _, err := StrictParseSymbol(k); err != nil {
				return fmt.Errorf("member key: %w", err)
			}

			if _, err := StrictValueFromProto(v); err != nil {
				return fmt.Errorf("member value: %w", err)
			}
		}
		return nil
	})(pb)
}

func NewStructValue(ctor Value, fields map[string]Value) Value {
	return MustNewValue(kittehs.Must1(StructValueFromProto(&valuesv1.Struct{
		Ctor: ctor.ToProto(),
		Fields: kittehs.TransformMapValues(
			fields,
			func(v Value) *valuesv1.Value { return v.ToProto() },
		),
	})))
}

func StructValueToStringMap(v Value) (map[string]Value, error) {
	_, fields := GetStructValue(v)
	return fields, nil
}

func FunctionValueFromProto(pb *valuesv1.Function) (FunctionValue, error) {
	return makeFromProto(func(pb *valuesv1.Function) error {
		if _, err := StrictParseExecutorID(pb.ExecutorId); err != nil {
			return fmt.Errorf("executor_id: %w", err)
		}

		if _, err := ParseSymbol(pb.Name); err != nil {
			return fmt.Errorf("name: %w", err)
		}

		return nil
	})(pb)
}

type FunctionFlag string

const (
	PrivilidgedFunctionFlag    FunctionFlag = "privilidged" // pass workflow context.
	PureFunctionFlag           FunctionFlag = "pure"        // do not run in an activity.
	DisablePollingFunctionFlag FunctionFlag = "no-poll"     // do not poll.
)

func (ff FunctionFlag) String() string { return string(ff) }

func NewFunctionValue(xid ExecutorID, name string, data []byte, flags []FunctionFlag, desc ModuleFunction) Value {
	return MustNewValue(kittehs.Must1(FunctionValueFromProto(&valuesv1.Function{
		ExecutorId: xid.String(),
		Name:       name,
		Desc:       desc.ToProto(),
		Data:       data,
		Flags:      kittehs.TransformToStrings(flags),
	})))
}
