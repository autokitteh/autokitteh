package sdktypes

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.jetify.com/typeid"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

const ValidIDChars = "0123456789abcdefghjkmnpqrstvwxyz"

type idTraits = typeid.PrefixType

type ID interface {
	json.Marshaler
	fmt.Stringer
	isValider
	stricter

	// Kind returns the type kind, which is the id prefix.
	Kind() string

	// Value returns the id value, meaning without the prefix.
	Value() *uuid.UUID

	UUIDValue() uuid.UUID

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
	return i.tid.Prefix()
}

func (i id[T]) Value() *uuid.UUID {
	if !i.IsValid() {
		return nil
	}

	u := uuid.UUID(i.tid.UUIDBytes())
	return &u
}

func (i id[T]) UUIDValue() uuid.UUID {
	if !i.IsValid() {
		return uuid.UUID{}
	}

	return uuid.UUID(i.tid.UUIDBytes())
}

func (i id[T]) UUIDValuePtr() *uuid.UUID {
	if !i.IsValid() {
		return nil
	}

	uuid := uuid.UUID(i.tid.UUIDBytes())
	return &uuid
}

func (i id[T]) MarshalJSON() ([]byte, error) {
	if !i.IsValid() {
		return []byte(`""`), nil
	}

	return json.Marshal(i.tid)
}

func (i *id[T]) UnmarshalJSON(data []byte) (err error) {
	if string(data) != `""` {
		err = json.Unmarshal(data, &i.tid)
	}

	return
}

func (i id[T]) ToTypeID() typeid.TypeID[T] { return i.tid }

func newID[ID id[T], T idTraits]() ID {
	uuid := NewUUID()
	return ID(id[T]{tid: typeid.Must(typeid.FromUUIDBytes[typeid.TypeID[T]](uuid[:]))})
}

func NewIDFromUUID[ID id[T], T idTraits](in uuid.UUID) (out ID) {
	if in != uuid.Nil {
		out = ID(id[T]{tid: typeid.Must(typeid.FromUUIDBytes[typeid.TypeID[T]](in[:]))})
	}
	return
}

func NewIDFromUUIDPtr[ID id[T], T idTraits](in *uuid.UUID) (_ ID) {
	if in == nil {
		return
	}

	return NewIDFromUUID[ID](*in)
}

func NewIDFromUUIDString[ID id[T], T idTraits](str string) (ID, error) {
	uuidUUID, err := uuid.Parse(str)
	if err != nil {
		return ID{}, err
	}
	return NewIDFromUUID[ID](uuidUUID), nil
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

func FromID[RetID id[T], T idTraits](id ID) RetID {
	return kittehs.Must1(ParseID[RetID](id.String()))
}

func IsIDOf[T idTraits](s string) bool {
	_, err := typeid.Parse[typeid.TypeID[T]](s)
	return err == nil
}

func IsID(s string) bool { return IsIDOf[typeid.AnyPrefix](s) }

func newNamedIDString(name, kind string) string {
	// a hint to the original name is encoded in the ID.
	idName := strings.Map(func(r rune) rune {
		if strings.ContainsRune(ValidIDChars, r) {
			return r
		}

		// Convert similar looking chars to something that is in the valid set.
		switch r {
		case 'i', 'l', '!':
			return '1'
		case 'u':
			return 'v'
		case 'o', '@':
			return '0'
		case '$':
			return 's'
		default:
			return '0'
		}
	}, name)

	if len(idName) > 6 {
		idName = idName[:6]
	}
	idName = fmt.Sprintf("%06s", idName)

	return fmt.Sprintf("%s_3kth%06s%016x", kind, idName, kittehs.HashString64(kind+","+name))
}
