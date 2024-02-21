package sdktypes

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	akproto "go.autokitteh.dev/autokitteh/proto"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
)

var protoMarshal = protojson.MarshalOptions{
	UseProtoNames: true,
}.Marshal

type Object interface {
	// all methods must be unexported.
	// we want all nil guards effective for outsiders.
	// this interface is just to form an algebric data type.

	isObject()
	toMessage() proto.Message
	toString() string
}

type comparableProto interface {
	comparable
	proto.Message
}

// object makes a protobuf present as an immutable object
// then can be extended upon.
type object[T comparableProto] struct {
	pb T // must never be nil.

	// TODO: this can actually alter the content of the object (see BuildResources),
	// rename to "prepare"?
	validatefn func(pb T) error // must never be nil.
}

func (*object[T]) isObject() {}

func ObjectToString(o Object) string {
	if o == nil {
		return ""
	}

	return o.toString()
}

func (o *object[T]) toString() string { return o.String() }

func (o *object[T]) String() string {
	if o == nil {
		return ""
	}

	txt, err := prototext.Marshal(o.pb)
	if err != nil {
		sdklogger.DPanic("prototext marshal error", "err", err)
		return "error"
	}

	return string(txt)
}

func (o *object[T]) toMessage() proto.Message { return proto.Clone(o.pb) }

func (o *object[T]) MarshalJSON() ([]byte, error) { return protoMarshal(o.pb) }

func (o *object[T]) UnmarshalJSON(data []byte) error {
	// OMG this is ugly, but works.
	pb := reflect.New(reflect.TypeOf(o.pb).Elem()).Interface().(T)

	if err := protojson.Unmarshal(data, pb); err != nil {
		return err
	}

	oo, err := fromProto(pb, o.validatefn)
	if err != nil {
		return err
	}

	*o = *oo

	return nil
}

func (o *object[T]) clone() *object[T] { return &object[T]{pb: proto.Clone(o.pb).(T)} }

func ToProto[T comparableProto](o *object[T]) T { return o.ToProto() }

func ToMessage(obj Object) proto.Message {
	if obj == nil {
		return nil
	}

	return obj.toMessage()
}

func (o *object[T]) ToProto() T {
	if o == nil {
		return *new(T) // nil for ptrs.
	}

	return o.clone().pb
}

func (o *object[T]) withValidator(v func(T) error) (*object[T], error) {
	oo := *o
	oo.validatefn = v

	if err := oo.validate(); err != nil {
		return nil, err
	}

	return &oo, nil
}

func (o *object[T]) validate() error {
	err := func() error {
		if proto.MessageName(o.pb) == "" || o.validatefn == nil {
			// must never happen.
			sdklogger.Panic("invalid object")
		}

		if err := akproto.Validate(o.pb); err != nil {
			return err
		}

		return o.validatefn(o.pb)
	}()
	if err != nil {
		return fmt.Errorf("%w: %w", sdkerrors.ErrInvalidArgument, err)
	}

	return nil
}

func (o *object[T]) Update(f func(T)) (*object[T], error) {
	return o.UpdateError(func(t T) error {
		f(t)
		return nil
	})
}

func (o *object[T]) UpdateError(f func(T) error) (*object[T], error) {
	if o == nil {
		sdklogger.DPanic("cannot update nil object")
		return nil, fmt.Errorf("cannot update nil object")
	}

	pb := o.clone().pb
	if err := f(pb); err != nil {
		return nil, err
	}

	return fromProto[T](pb, o.validatefn)
}

func Equal(a, b Object) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return proto.Equal(a.toMessage(), b.toMessage())
}

// Initializes object with a given pb. If pb is nil, nil is returned.
// v is the validator set on the object.
func fromProto[T comparableProto](pb T, v func(T) error,
) (*object[T], error) {
	var zero T // nil for ptrs
	if pb == zero {
		return nil, nil
	}

	if v == nil {
		v = func(T) error { return nil }
	}

	if err := v(pb); err != nil {
		return nil, err
	}

	return &object[T]{pb: pb, validatefn: v}, nil
}

func makeMustFromProto[T comparableProto](v func(T) error) func(T) *object[T] {
	return func(pb T) *object[T] { return kittehs.Must1(fromProto(pb, v)) }
}

func makeFromProto[T comparableProto](v func(T) error) func(T) (*object[T], error) {
	return func(pb T) (*object[T], error) { return fromProto(pb, v) }
}

func makeWithValidator[T comparableProto](v func(T) error) func(o *object[T]) (*object[T], error) {
	return func(o *object[T]) (*object[T], error) { return o.withValidator(v) }
}

func GetObjectHash(o Object) string {
	if o == nil {
		return ""
	}

	hash := sha512.Sum512_256(kittehs.Must1(proto.Marshal(o.toMessage())))
	return hex.EncodeToString(hash[:])
}
