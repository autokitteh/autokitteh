package sdktypes

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	akproto "go.autokitteh.dev/autokitteh/proto"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type Object interface {
	json.Marshaler
	fmt.Stringer
	isValider
	stricter

	isObject()
}

type objectTraits[M interface{ proto.Message }] interface {
	// Validate is used to validate all fields in the given
	// message. If a field is missing, it can be ignored.
	Validate(m M) error

	// StrictValidate is used to validate that all mandatory
	// fields are specified. It does not need to call Validate,
	// the underlying object will do it.
	StrictValidate(m M) error
}

type nopObjectTraits[M proto.Message] struct{}

func (nopObjectTraits[M]) Validate(m M) error       { return nil }
func (nopObjectTraits[M]) StrictValidate(m M) error { return nil }

var _ objectTraits[proto.Message] = nopObjectTraits[proto.Message]{}

type comparableMessage interface {
	proto.Message
	comparable // for comparison to nil (proto.Message will always be a ptr).
}

type object[M comparableMessage, T objectTraits[M]] struct {
	kittehs.DoNotCompare

	m M
}

func clone[M proto.Message](m M) M { return proto.Clone(m).(M) }

func (o object[T, M]) isObject()              {}
func (o object[M, T]) IsValid() bool          { var zero M; return o.m != zero }
func (o object[M, T]) ToProto() M             { return clone(o.m) }
func (o object[M, T]) Message() proto.Message { return o.ToProto() }
func (o object[M, T]) IsZero() bool           { return proto.Size(o.m) == 0 }

// the returned message will not always be the message stored in the object.
func (o *object[M, T]) read() M {
	if o.IsValid() {
		return o.m
	}

	return reflect.New(reflect.TypeOf(o.m).Elem()).Interface().(M)
}

// sets the message is nil. this mutates the object.
func (o *object[M, T]) reset() { var zero M; o.m = zero }

func (o object[M, T]) String() string {
	if !o.IsValid() {
		return ""
	}

	return string(kittehs.Must1(prototext.Marshal(o.m)))
}

// forceUpdate replaces the message without validation.
// This can be called only and only if the message is known to be valid.
func (o object[M, T]) forceUpdate(f func(M)) object[M, T] {
	m := proto.Clone(o.read()).(M)
	f(m)
	return object[M, T]{m: m}
}

var protoMarshal = protojson.MarshalOptions{UseProtoNames: true}.Marshal

func (o object[M, T]) MarshalJSON() ([]byte, error) {
	// The object can be marshalled as a pointer, so if it's null, we just
	// specify null in JSON.
	if !o.IsValid() {
		return []byte("null"), nil
	}
	return protoMarshal(o.m)
}

func (o *object[M, T]) UnmarshalJSON(b []byte) (err error) {
	// The object can be marshalled as a pointer, so if it's null, we reset the object.
	// (ie. we got an invalid/nil object)
	if string(b) == "null" {
		o.reset()
		return
	}

	o.m = o.read()

	if err = protojson.Unmarshal(b, o.m); err != nil {
		return
	}

	if err = validate[M, T](o.m); err != nil {
		o.reset()
		return
	}

	return
}

func (o object[M, T]) Strict() error {
	if !o.IsValid() {
		return sdkerrors.NewInvalidArgumentError("zero object")
	}

	var t T
	return t.StrictValidate(o.m)
}

func (o object[M, T]) Hash() string { return hash(o.m) }

func (o object[M, T]) Equal(other interface{ ToProto() M }) bool {
	return proto.Equal(o.m, other.ToProto())
}

func strictValidate[M proto.Message, T objectTraits[M]](m M) error {
	var zero M
	if proto.Equal(zero, m) {
		return errors.New("empty")
	}

	var t T
	if err := t.StrictValidate(m); err != nil {
		return sdkerrors.ErrInvalidArgument{Underlying: err}
	}

	return validate[M, T](m)
}

func validate[M proto.Message, T objectTraits[M]](m M) error {
	var zero M
	if proto.Equal(zero, m) {
		return nil
	}

	if err := akproto.Validate(m); err != nil {
		return sdkerrors.ErrInvalidArgument{Underlying: err}
	}

	var t T
	if err := t.Validate(m); err != nil {
		return sdkerrors.ErrInvalidArgument{Underlying: err}
	}

	return nil
}

func fromProto[M comparableMessage, T objectTraits[M]](m M) (o object[M, T], err error) {
	if err = validate[M, T](m); err != nil {
		return
	}

	o = object[M, T]{m: clone(m)}
	return
}

func forceFromProto[W ~struct{ object[M, T] }, M comparableMessage, T objectTraits[M]](m M) W {
	var zero M
	if proto.Equal(m, zero) {
		return W{}
	}
	return W{object[M, T]{m: clone(m)}}
}

func FromProto[W ~struct{ object[M, T] }, M comparableMessage, T objectTraits[M]](m M) (w W, err error) {
	var o object[M, T]
	if o, err = fromProto[M, T](m); err != nil {
		return
	}

	w = W{o}
	return
}

// Use this to create a valid, but empty object.
func zeroObject[W ~struct{ object[M, T] }, M comparableMessage, T objectTraits[M]]() W {
	o := object[M, T]{}
	o.m = o.read()
	return W{o}
}

func ToProto[O interface{ ToProto() M }, M proto.Message](o O) M { return o.ToProto() }
