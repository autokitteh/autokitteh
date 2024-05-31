package sdktypes

import (
	"errors"
	"time"

	"golang.org/x/exp/constraints"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	valuev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
)

// ---

type NothingValuePB = valuev1.Nothing

type NothingValue struct {
	object[*NothingValuePB, nopObjectTraits[*NothingValuePB]]
}

func (NothingValue) isConcreteValue() {}

var Nothing = forceFromProto[Value](&ValuePB{Nothing: &NothingValuePB{}})

func (v Value) IsNothing() bool          { return v.read().Nothing != nil }
func (v Value) GetNothing() NothingValue { return forceFromProto[NothingValue](v.read().Nothing) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsNothing() {
			return v.GetNothing()
		}
		return nil
	})
}

// ---

type SymbolValuePB = valuev1.Symbol

type symbolValueTraits struct{}

func (symbolValueTraits) Validate(m *SymbolValuePB) error {
	return symbolTraits{}.Validate(m.Name)
}

func (symbolValueTraits) StrictValidate(m *SymbolValuePB) error {
	if m.Name == "" {
		return errors.New("empty")
	}

	return nil
}

var _ objectTraits[*SymbolValuePB] = symbolValueTraits{}

func (SymbolValue) isConcreteValue() {}

type SymbolValue struct {
	object[*SymbolValuePB, symbolValueTraits]
}

func (s SymbolValue) Symbol() Symbol { return NewSymbol(s.read().Name) }

func (v Value) IsSymbol() bool         { return v.read().Symbol != nil }
func (v Value) GetSymbol() SymbolValue { return forceFromProto[SymbolValue](v.read().Symbol) }

func NewSymbolValue(s Symbol) Value {
	return forceFromProto[Value](&ValuePB{Symbol: &SymbolValuePB{Name: string(s.String())}})
}

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsSymbol() {
			return v.GetSymbol()
		}
		return nil
	})
}

// ---

type StringValuePB = valuev1.String

type StringValue struct {
	object[*StringValuePB, nopObjectTraits[*StringValuePB]]
}

func (StringValue) isConcreteValue() {}

func (s StringValue) Value() string { return s.read().V }

func NewStringValue(s string) Value {
	return forceFromProto[Value](&ValuePB{String_: &StringValuePB{V: s}})
}

func (v Value) IsString() bool         { return v.read().String_ != nil }
func (v Value) GetString() StringValue { return forceFromProto[StringValue](v.read().String_) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsString() {
			return v.GetString()
		}
		return nil
	})
}

// ---

type IntegerValuePB = valuev1.Integer

type IntegerValue struct {
	object[*IntegerValuePB, nopObjectTraits[*IntegerValuePB]]
}

func (IntegerValue) isConcreteValue() {}

func (s IntegerValue) Value() int64 { return s.read().V }

func NewIntegerValue[T constraints.Integer](v T) Value {
	return forceFromProto[Value](&ValuePB{Integer: &IntegerValuePB{V: int64(v)}})
}

func (v Value) IsInteger() bool          { return v.read().Integer != nil }
func (v Value) GetInteger() IntegerValue { return forceFromProto[IntegerValue](v.read().Integer) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsInteger() {
			return v.GetInteger()
		}
		return nil
	})
}

// ---

type BooleanValuePB = valuev1.Boolean

type BooleanValue struct {
	object[*BooleanValuePB, nopObjectTraits[*BooleanValuePB]]
}

func (BooleanValue) isConcreteValue() {}

func (s BooleanValue) Value() bool { return s.read().V }

func NewBooleanValue(v bool) Value {
	return forceFromProto[Value](&ValuePB{Boolean: &BooleanValuePB{V: v}})
}

func (v Value) IsBoolean() bool          { return v.read().Boolean != nil }
func (v Value) GetBoolean() BooleanValue { return forceFromProto[BooleanValue](v.read().Boolean) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsBoolean() {
			return v.GetBoolean()
		}
		return nil
	})
}

// ---

type FloatValuePB = valuev1.Float

type FloatValue struct {
	object[*FloatValuePB, nopObjectTraits[*FloatValuePB]]
}

func (FloatValue) isConcreteValue() {}

func (s FloatValue) Value() float64 { return s.read().V }

func NewFloatValue[T constraints.Float](v T) Value {
	return forceFromProto[Value](&ValuePB{Float: &FloatValuePB{V: float64(v)}})
}

func (v Value) IsFloat() bool        { return v.read().Float != nil }
func (v Value) GetFloat() FloatValue { return forceFromProto[FloatValue](v.read().Float) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsFloat() {
			return v.GetFloat()
		}
		return nil
	})
}

// ---

type DurationValuePB = valuev1.Duration

type DurationValue struct {
	object[*DurationValuePB, nopObjectTraits[*DurationValuePB]]
}

func (DurationValue) isConcreteValue() {}

func (s DurationValue) Value() time.Duration { return s.read().V.AsDuration() }

func NewDurationValue(v time.Duration) Value {
	return forceFromProto[Value](&ValuePB{Duration: &DurationValuePB{V: durationpb.New(v)}})
}

func (v Value) IsDuration() bool           { return v.read().Duration != nil }
func (v Value) GetDuration() DurationValue { return forceFromProto[DurationValue](v.read().Duration) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsDuration() {
			return v.GetDuration()
		}
		return nil
	})
}

// ---

type TimeValuePB = valuev1.Time

type TimeValue struct {
	object[*TimeValuePB, nopObjectTraits[*TimeValuePB]]
}

func (TimeValue) isConcreteValue() {}

func (s TimeValue) Value() time.Time { return s.read().V.AsTime() }

func NewTimeValue(v time.Time) Value {
	return forceFromProto[Value](&ValuePB{Time: &TimeValuePB{V: timestamppb.New(v)}})
}

func (v Value) IsTime() bool       { return v.read().Time != nil }
func (v Value) GetTime() TimeValue { return forceFromProto[TimeValue](v.read().Time) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsTime() {
			return v.GetTime()
		}
		return nil
	})
}
