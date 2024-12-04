package sdktypes

import (
	"errors"
	"fmt"
	"net/url"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	valuev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
)

func mandatory[T comparable](name string, t T) error {
	var zero T
	if t == zero {
		return fmt.Errorf("%s: missing", name)
	}
	return nil
}

func mandatorySlice[T any](name string, t []T) error {
	if len(t) == 0 {
		return fmt.Errorf("%s: missing", name)
	}
	return nil
}

func nonzeroMessage(m proto.Message) error {
	isZero := true

	m.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		isZero = false
		return false
	})

	if isZero {
		return errors.New("empty")
	}

	return nil
}

func oneOfMessage(m proto.Message, ignores ...string) error {
	var count int

	isIgnored := kittehs.ContainedIn(ignores...)

	m.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if isIgnored(string(fd.Name())) {
			return true
		}

		count++
		return count < 2
	})

	if count != 1 {
		return errors.New("exactly one field must be set")
	}

	return nil
}

func indexedObject[O ~struct{ object[M, T] }, M comparableMessage, T objectTraits[M]](i int, m M) error {
	if err := strictValidate[M, T](m); err != nil {
		return errorForValue(i, err)
	}
	return nil
}

func objectField[_ ~struct{ object[M, T] }, M comparableMessage, T objectTraits[M]](name string, m M) error {
	if err := validate[M, T](m); err != nil {
		return errorForValue(name, err)
	}
	return nil
}

func objectsSlice[O ~struct{ object[M, T] }, M comparableMessage, T objectTraits[M]](ms []M) error {
	return errorForValue(kittehs.ValidateList(ms, indexedObject[O, M, T]))
}

func objectsSliceField[O ~struct{ object[M, T] }, M comparableMessage, T objectTraits[M]](name string, ms []M) error {
	err := errorForValue(kittehs.ValidateList(ms, indexedObject[O, M, T]))
	if err != nil {
		return errorForValue(name, err)
	}
	return nil
}

func varScopeIDField(s string) error {
	if _, err := ParseVarScopeID(s); err != nil {
		return errorForValue("scope_id", err)
	}
	return nil
}

func idField[ID id[T], T idTraits](name string, s string) error {
	if _, err := ParseID[ID](s); err != nil {
		return errorForValue(name, err)
	}
	return nil
}

func valuesSlice(vs []*valuev1.Value) error { return objectsSlice[Value](vs) }
func valuesSliceField(name string, vs []*valuev1.Value) error {
	return objectsSliceField[Value](name, vs)
}

func objectsMapField[O ~struct{ object[M, T] }, M comparableMessage, T objectTraits[M]](name string, m map[string]M) error {
	keys := make(map[string]bool, len(m))

	err := errors.Join(kittehs.TransformMapToList(m, func(k string, v M) error {
		if keys[k] {
			return fmt.Errorf("duplicate key %q", k)
		}

		if _, err := Strict(ParseSymbol(k)); err != nil {
			return fmt.Errorf("name %q: %w", k, err)
		}

		keys[k] = true

		if err := strictValidate[M, T](v); err != nil {
			return fmt.Errorf("value %q: %w", k, err)
		}

		return nil
	})...)

	if err == nil {
		return nil
	}

	return errorForValue(name, err)
}

func valuesMapField(name string, m map[string]*valuev1.Value) error {
	return objectsMapField[Value](name, m)
}

func nameField(name string, s string) error {
	if _, err := ParseSymbol(s); err != nil {
		return errorForValue(name, err)
	}
	return nil
}

func executorIDField(name string, s string) error {
	if _, err := ParseExecutorID(s); err != nil {
		return errorForValue(name, err)
	}
	return nil
}

func enumField[W ~struct{ enum[T, E] }, T enumTraits, E ~int32](name string, v E) error {
	if _, err := EnumFromProto[W, T, E](v); err != nil {
		return errorForValue(name, err)
	}
	return nil
}

func urlField(name string, s string) error {
	if s == "" {
		return nil
	}

	if _, err := url.Parse(s); err != nil {
		return errorForValue(name, err)
	}
	return nil
}

func symbolField(name string, s string) error {
	if _, err := ParseSymbol(s); err != nil {
		return errorForValue(name, err)
	}
	return nil
}

// errorForValue wraps the given error by appending the value to
// it. If the error is nil, it also returns nil. The error comes
// before the value and not after, because it describes what's
// wrong with the value. This allows additional error wrapping.
// Example: "loop error: counter must be positive: -1".
func errorForValue[T any](value T, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", err, value)
}
