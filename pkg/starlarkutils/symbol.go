package starlarkutils

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Adapted from https://github.com/google/starlark-go/blob/7a1108eaa0124ea025e567d9feb4e41aab4bb024/starlarkstruct/struct_test.go#L48.

// A symbol is a distinct value that acts as a constructor of "branded"
// struct instances, like a class symbol in Python or a "provider" in Bazel.
type Symbol string

var _ starlark.Callable = Symbol("")

func (sym Symbol) Name() string          { return string(sym) }
func (sym Symbol) String() string        { return string(sym) }
func (sym Symbol) Type() string          { return "symbol" }
func (sym Symbol) Freeze()               {} // immutable
func (sym Symbol) Truth() starlark.Bool  { return starlark.True }
func (sym Symbol) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable: %s", sym.Type()) }

func (sym Symbol) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("%s: unexpected positional arguments", sym)
	}
	return starlarkstruct.FromKeywords(sym, kwargs), nil
}

func GenSymbol(_ *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "name", &name); err != nil {
		return nil, err
	}
	return Symbol(name), nil
}
