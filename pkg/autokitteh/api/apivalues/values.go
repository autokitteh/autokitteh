package apivalues

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNoConversion = errors.New("no convertion")

type NoneValue struct{}

func (NoneValue) String() string { return "<none>" }

func (NoneValue) isValue() {}

var None = MustNewValue(NoneValue{})

func (NoneValue) ConvertFrom(v value) error {
	switch v.(type) {
	case NoneValue:
		return nil
	default:
		return ErrNoConversion
	}
}

//--

type StringValue string

func (s StringValue) String() string { return (string)(s) }

func (StringValue) isValue() {}

func String(s string) *Value { return MustNewValue(StringValue(s)) }

func (s StringValue) UnwrapInto(dst interface{}) (bool, error) {
	switch v := dst.(type) {
	case *time.Duration:
		d, err := time.ParseDuration(string(s))
		if err != nil {
			return false, err
		}

		*v = d
		return true, nil

	default:
		return false, nil
	}
}

func (s *StringValue) ConvertFrom(v value) error {
	switch v := v.(type) {
	case StringValue:
		*s = v
		return nil
	default:
		return ErrNoConversion
	}
}

//--

type SymbolValue string

func (s SymbolValue) String() string { return (string)(s) }

func (SymbolValue) isValue() {}

func Symbol(s string) *Value { return MustNewValue(SymbolValue(s)) }

func (s *SymbolValue) ConvertFrom(v value) error {
	switch v := v.(type) {
	case SymbolValue:
		*s = v
		return nil
	default:
		return ErrNoConversion
	}
}

//--

type IntegerValue int64

func (i IntegerValue) String() string { return fmt.Sprintf("%d", i) }

func (IntegerValue) isValue() {}

func Integer(i int64) *Value { return MustNewValue(IntegerValue(i)) }

func (i IntegerValue) UnwrapInto(dst interface{}) (bool, error) {
	switch v := dst.(type) {
	case *time.Time:
		*v = time.Unix(int64(i), 0)
		return true, nil

	case *time.Duration:
		*v = time.Duration(int64(i)) * time.Second
		return true, nil

	default:
		return false, nil
	}
}

func (i *IntegerValue) ConvertFrom(v value) error {
	switch v := v.(type) {
	case IntegerValue:
		*i = v
		return nil

	case FloatValue:
		if !v.IsIntegral() {
			return errors.New("float is not integeral")
		}

		*i = IntegerValue(int64(v))
		return nil
	default:
		return ErrNoConversion
	}
}

//--

type BooleanValue bool

func (b BooleanValue) String() string { return fmt.Sprintf("%t", b) }

func (BooleanValue) isValue() {}

func Boolean(b bool) *Value { return MustNewValue(BooleanValue(b)) }

func (b *BooleanValue) ConvertFrom(v value) error {
	switch v := v.(type) {
	case BooleanValue:
		*b = v
		return nil
	default:
		return ErrNoConversion
	}
}

//--

type FloatValue float32

func (f FloatValue) String() string { return fmt.Sprintf("%f", f) }

func (FloatValue) isValue() {}

func Float(f float32) *Value { return MustNewValue(FloatValue(f)) }

func (f FloatValue) IsIntegral() bool { return float64(f) == float64(int64(f)) }

func (f FloatValue) UnwrapInto(dst interface{}) (bool, error) {
	switch v := dst.(type) {
	case *time.Duration:
		*v = time.Duration(float32(f) * float32(time.Second))
		return true, nil

	default:
		return false, nil
	}
}

func (f *FloatValue) ConvertFrom(v value) error {
	switch v := v.(type) {
	case FloatValue:
		*f = v
		return nil
	case IntegerValue:
		*f = FloatValue(v)
		return nil

	default:
		return ErrNoConversion
	}
}

//--

type ListValue []*Value

func (l ListValue) String() string {
	vs := make([]string, len(l))
	for i, v := range l {
		vs[i] = v.String()
	}
	return strings.Join(vs, " ,")
}

func (ListValue) isValue() {}

func List(items ...*Value) *Value { return MustNewValue(ListValue(items)) }

func StringList(items ...string) *Value {
	vs := make([]*Value, len(items))
	for i, s := range items {
		vs[i] = String(s)
	}
	return List(vs...)
}

func (l *ListValue) ConvertFrom(v value) error {
	switch v := v.(type) {
	case ListValue:
		*l = v
		return nil

	case SetValue:
		*l = ListValue(v)
		return nil

	default:
		return ErrNoConversion
	}
}

func GetListValue(v *Value) ListValue {
	if vv, ok := v.Get().(ListValue); ok {
		return vv
	}

	return nil
}

//--

type DictItem struct{ K, V *Value }

type DictValue []*DictItem

func (d DictValue) String() string {
	vs := make([]string, 0, len(d))
	for _, v := range d {
		vs = append(vs, fmt.Sprintf("%v: %v", v.K, v.V))
	}
	return strings.Join(vs, " ,")
}

func (d DictValue) ToStringValuesMap(dst map[string]*Value) {
	for _, i := range d {
		dst[i.K.Get().String()] = i.V
	}
}

func (DictValue) isValue() {}

func Dict(items ...*DictItem) *Value { return MustNewValue(DictValue(items)) }

func DictFromMap(m map[string]*Value) *Value {
	items := make([]*DictItem, 0, len(m))
	for k, v := range m {
		items = append(items, &DictItem{K: String(k), V: v})
	}
	return Dict(items...)
}

func (d DictValue) GetKey(k *Value) *DictItem {
	for _, i := range d {
		if i.K.Equal(k) {
			return i
		}
	}

	return nil
}

func (d *DictValue) ConvertFrom(v value) error {
	switch v := v.(type) {
	case DictValue:
		*d = v
		return nil

	default:
		return ErrNoConversion
	}
}

func GetDictValue(v *Value) *DictValue { return GetConcretValue[DictValue](v) }

//--

type BytesValue []byte

func (b BytesValue) String() string { return fmt.Sprintf("%q", ([]byte)(b)) }

func (BytesValue) isValue() {}

func Bytes(bs []byte) *Value { return MustNewValue(BytesValue(bs)) }

func (b *BytesValue) ConvertFrom(v value) error {
	switch v := v.(type) {
	case BytesValue:
		*b = v
		return nil

	default:
		return ErrNoConversion
	}
}

//--

type TimeValue time.Time

func (t TimeValue) String() string { return time.Time(t).String() }

func (TimeValue) isValue() {}

func Time(t time.Time) *Value { return MustNewValue(TimeValue(t)) }

func (t TimeValue) UnwrapInto(dst interface{}) (bool, error) {
	switch v := dst.(type) {
	case *time.Time:
		*v = time.Time(t)
		return true, nil

	case *string:
		*v = t.String()
		return true, nil

	case *int64:
		*v = time.Time(t).Unix()
		return true, nil

	default:
		return false, nil
	}
}

func (t *TimeValue) ConvetFrom(v value) error {
	switch v := v.(type) {
	case TimeValue:
		*t = v
		return nil

	case IntegerValue:
		*t = TimeValue(time.Unix(int64(v), 0))
		return nil

	default:
		return ErrNoConversion
	}
}

//--

type DurationValue time.Duration

func (d DurationValue) String() string { return time.Duration(d).String() }

func (DurationValue) isValue() {}

func Duration(d time.Duration) *Value { return MustNewValue(DurationValue(d)) }

func (d DurationValue) UnwrapInto(dst interface{}) (bool, error) {
	switch v := dst.(type) {
	case *time.Duration:
		*v = time.Duration(d)
		return true, nil

	case *string:
		*v = d.String()
		return true, nil

	default:
		return false, nil
	}
}

func (d *DurationValue) ConvertFrom(v value) error {
	switch v := v.(type) {
	case DurationValue:
		*d = v
		return nil

	case IntegerValue:
		*d = DurationValue(time.Duration(v) * time.Second)
		return nil

	case FloatValue:
		*d = DurationValue(float64(v) * float64(time.Second))
		return nil

	case StringValue:
		dd, err := time.ParseDuration(string(v))
		if err != nil {
			return err
		}

		*d = DurationValue(dd)
		return nil

	default:
		return ErrNoConversion
	}
}

//--

type SetValue []*Value

func (s SetValue) String() string {
	vs := make([]string, len(s))
	for i, v := range s {
		vs[i] = v.String()
	}
	return strings.Join(vs, " ,")
}

func (SetValue) isValue() {}

func Set(vs ...*Value) *Value { return MustNewValue(SetValue(vs)) }

func (s *SetValue) ConvertFrom(v value) error {
	switch v := v.(type) {
	case SetValue:
		*s = v
		return nil

	default:
		return ErrNoConversion
	}
}

//--

type CallValue struct {
	Name   string          `json:"name"`
	ID     string          `json:"id"`
	Issuer string          `json:"issuer"`
	Flags  map[string]bool `json:"flags"` // set only by ak. not by user.
}

func (c CallValue) String() string {
	return fmt.Sprintf("call %q/%s/%s", c.Name, c.ID, c.Issuer)
}

func (CallValue) isValue() {}

func GetCallValue(v *Value) *CallValue { return GetConcretValue[CallValue](v) }

//--

type StructValue struct {
	Ctor   *Value            `json:"ctor"`
	Fields map[string]*Value `json:"fields"`
}

func (s StructValue) String() string { return fmt.Sprintf("struct %v", s.Ctor) } // TODO: add fields.

func (StructValue) isValue() {}

func Struct(ctor *Value, fields map[string]*Value) *Value {
	return MustNewValue(StructValue{Ctor: ctor, Fields: fields})
}

func GetStructValue(v *Value) *StructValue { return GetConcretValue[StructValue](v) }

//--

type ModuleValue struct {
	Name    string            `json:"name"`
	Members map[string]*Value `json:"members"`
}

func (c ModuleValue) String() string { return fmt.Sprintf("module %v", c.Name) }

func (ModuleValue) isValue() {}

func Module(name string, members map[string]*Value) *Value {
	return MustNewValue(ModuleValue{Name: name, Members: members})
}

func GetModuleValue(v *Value) *ModuleValue { return GetConcretValue[ModuleValue](v) }

//--

type FunctionSignature struct {
	Name          string   `json:"name"`
	Doc           string   `json:"doc"`
	NumArgs       uint32   `json:"num_args"`
	NumKWOnlyArgs uint32   `json:"num_kwonly_args"`
	ArgsNames     []string `json:"args_names"`
	HasKWArgs     bool     `json:"has_kwargs"`
	HasVarargs    bool     `json:"has_varargs"`
}

type FunctionValue struct {
	Lang   string `json:"lang"`
	FuncID string `json:"func_id"`
	Scope  string `json:"scope"`

	Signature *FunctionSignature `json:"signature"` // optional
}

func (c FunctionValue) String() string {
	return fmt.Sprintf("%s.%s.%s", c.FuncID, c.Lang, c.Scope)
}

func ParseFunctionValueString(s string) (*FunctionValue, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid function value")
	}

	return &FunctionValue{
		FuncID: parts[0],
		Lang:   parts[1],
		Scope:  parts[2],
	}, nil
}

func (FunctionValue) isValue() {}

func GetFunctionValue(v *Value) *FunctionValue { return GetConcretValue[FunctionValue](v) }
