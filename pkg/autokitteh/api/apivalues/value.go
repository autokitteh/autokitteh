package apivalues

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	pbvalues "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/values"
)

type value interface {
	fmt.Stringer
	isValue()
}

type converter interface{ ConvertFrom(value) error }

type Value struct{ pb *pbvalues.Value }

func (v *Value) PB() *pbvalues.Value {
	if v == nil || v.pb == nil {
		return nil
	}

	return proto.Clone(v.pb).(*pbvalues.Value)
}

func (v *Value) Clone() *Value { return &Value{pb: v.PB()} }

func (v *Value) String() string { return v.Get().String() }

func ValueFromProto(pb *pbvalues.Value) (*Value, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	return (&Value{pb: pb}).Clone(), nil
}

func MustValueFromProto(pb *pbvalues.Value) *Value {
	v, err := ValueFromProto(pb)
	if err != nil {
		panic(fmt.Errorf("%w: %v", err, pb))
	}
	return v
}

func NewValue(v value) (*Value, error) { return ValueFromProto(toProto(v)) }

func MustNewValue(v value) *Value { return MustValueFromProto(toProto(v)) }

func (v *Value) Equal(o *Value) bool { return proto.Equal(v.PB(), o.PB()) }

func (v *Value) Get() value {
	if v == nil || v.pb == nil {
		return NoneValue{}
	}

	ret, err := fromProto(v.pb.Type)
	if err != nil {
		panic(err)
	}

	return ret
}

func (v *Value) set(vv value) { v.pb = toProto(vv) }

func (v *Value) Unwrap(fopts ...func(*unwrapOpts)) interface{} { return Unwrap(v.Get(), fopts...) }

func (v *Value) IsEphemeral() bool {
	switch vv := v.Get().(type) {
	case NoneValue, StringValue, SymbolValue, IntegerValue, BooleanValue, FloatValue, BytesValue,
		TimeValue, DurationValue:
		return false
	case CallValue, FunctionValue:
		return true
	case SetValue:
		for _, i := range vv {
			if i.IsEphemeral() {
				return true
			}
		}
		return false

	case ListValue:
		for _, i := range vv {
			if i.IsEphemeral() {
				return true
			}
		}
		return false

	case DictValue:
		for _, i := range vv {
			if i.K.IsEphemeral() || i.V.IsEphemeral() {
				return true
			}
		}
		return false

	case StructValue:
		if vv.Ctor.IsEphemeral() {
			return true
		}

		for _, f := range vv.Fields {
			if f.IsEphemeral() {
				return true
			}
		}
		return false

	case ModuleValue:
		for _, m := range vv.Members {
			if m.IsEphemeral() {
				return true
			}
		}
		return false

	default:
		panic("unhandled type")
	}
}

func GetConcretValue[T any](v *Value) *T {
	if vv, ok := v.Get().(T); ok {
		return &vv
	}

	return nil
}
