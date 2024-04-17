package sdktypes

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.jetpack.io/typeid"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

const validIDChars = "0123456789abcdefghjkmnpqrstvwxyz"

type idTraits = typeid.PrefixType

type ID interface {
	json.Marshaler
	fmt.Stringer
	isValider
	stricter

	// Kind returns the type kind, which is the id prefix.
	Kind() string

	// Value returns the id value, meaning without the prefix.
	Value() string

	UUIDValue() *UUID

	isID()
}

// TypeID is not embedded in order not to expose it outside.
// Its behaviour is not exactly what we want, eg. String() on
// its zero value will return a zeroed out id instead of just
// the empty string.
// TODO: Replace typeid with our own implementation.
type id[T idTraits] struct{ tid typeid.TypeID[T] }

func (i id[T]) isID()         {}
func (i id[T]) IsValid() bool { var zero id[T]; return i != zero }
func (i id[T]) Hash() string  { return hash(wrapperspb.String(i.String())) }

func (i id[T]) Strict() error {
	if !i.IsValid() {
		return sdkerrors.NewInvalidArgumentError("invalid")
	}

	return nil
}

func (i id[T]) String() string {
	if !i.IsValid() {
		return ""
	}

	return i.tid.String()
}

func (i id[T]) Kind() string {
	if !i.IsValid() {
		return ""
	}

	return i.tid.Prefix()
}

func (i id[T]) Value() string {
	if !i.IsValid() {
		return ""
	}

	return i.tid.Suffix()
}

func (i id[T]) UUIDValue() *UUID {

	if !i.IsValid() {
		return nil
	}

	u := uuid.UUID(i.tid.UUIDBytes())
	return &u
}

func (i id[T]) MarshalJSON() ([]byte, error)           { return json.Marshal(i.tid) }
func (i *id[T]) UnmarshalJSON(data []byte) (err error) { err = json.Unmarshal(data, &i.tid); return }

func newID[ID id[T], T idTraits]() ID {
	var t T
	tid := kittehs.Must1(typeid.FromUUIDBytes[typeid.TypeID[T]](t.Prefix(), newUUID()))

	return ID(id[T]{tid: tid})
}

func NewIDFromUUID[ID id[T], T idTraits](uuid *UUID) ID {
	if uuid == nil {
		var zero ID
		return zero
	}
	var t T
	tid := kittehs.Must1(typeid.FromUUIDBytes[typeid.TypeID[T]](t.Prefix(), uuid[:]))

	return ID(id[T]{tid: tid})
}

func ParseID[ID id[T], T idTraits](s string) (ID, error) {
	var zero ID

	if s == "" {
		return zero, nil
	}

	tid, err := typeid.Parse[typeid.TypeID[T]](s)
	if err != nil {
		return zero, err
	}

	return ID(id[T]{tid: tid}), nil
}

func IsIDOf[T idTraits](s string) bool {
	_, err := typeid.Parse[typeid.TypeID[T]](s)
	return err == nil
}

func IsID(s string) bool { return IsIDOf[typeid.AnyPrefix](s) }
