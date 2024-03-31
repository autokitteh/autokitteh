package sdktypes

import (
	"github.com/iancoleman/strcase"
)

// CAVEATS:
// - All integers are treated as int64.
// - All floats are treated as float64.
// - json.Number is converted to either int64 or float64, depends on its value. If neither matches, as a string.
// - json.RawMessage is un/marshalled directly into/from Value.
// - functions that are wrapped by a specific wrapper can be unwrapped only with that specific wrapper instance.
type ValueWrapper struct {
	// Wrap structs as maps.
	WrapStructAsMap bool

	// Unwrap: Treat all dict keys as strings (JSON does not deal well with map[any]any).
	// TODO: maybe allow ints, floats, any hashable as well somehow.
	SafeForJSON bool

	// Wrap: Used for functions that are wrapped using this wrapper.
	ExecutorID ExecutorID

	FromStructFieldNameCaser func(string) string
	ToStructFieldNameCaser   func(string) string

	// Unwrap: if a struct, marshal it to JSON directly, do not convert to map.
	UnwrapStructsAsJSON bool

	// Error out when trying to unwrap into a struct and the struct has fields that do not exist in the value.
	UnwrapErrorOnNonexistentStructFields bool

	// Unwrap: transform duration into microseconds, do not convert to string.
	RawDuration bool

	// Unwrap: Tranform value before unwrapping. If returns InvalidValue, ignore value.
	Preunwrap func(Value) (Value, error)

	// Unwrap: if not handled, use this unwrapper.
	UnwrapUnknown func(Value) (any, error)
}

var DefaultValueWrapper ValueWrapper

const ctorFieldName = "__ctor"

func (w ValueWrapper) fromStructCaser(x string) string {
	if w.FromStructFieldNameCaser == nil {
		return strcase.ToSnake(x)
	}

	return w.FromStructFieldNameCaser(x)
}

func (w ValueWrapper) toStructCaser(x string) string {
	if w.ToStructFieldNameCaser == nil {
		return strcase.ToCamel(x)
	}

	return w.ToStructFieldNameCaser(x)
}
