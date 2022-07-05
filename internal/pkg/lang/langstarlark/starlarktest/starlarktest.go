// Adapted from https://github.com/google/starlark-go/blob/master/starlarktest/starlarktest.go.
package starlarktest

import (
	"fmt"
	"strings"

	"go.starlark.net/starlark"
)

var FailBuiltin = starlark.NewBuiltin("fail", fail)
var AssertBuiltin = starlark.NewBuiltin("assert", assert)
var CatchBuiltin = starlark.NewBuiltin("catch", catch)

// returns tuple (retval, err_as_string):
// - if succeeds, returns (retval, None)
// - if fails, returns (None, err_as_string)
func catch(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var fn starlark.Callable
	if err := starlark.UnpackArgs("catch", args, kwargs, "fn", &fn); err != nil {
		return nil, err
	}
	ret, err := starlark.Call(thread, fn, nil, nil)
	if err != nil {
		return starlark.Tuple(
			[]starlark.Value{
				starlark.None,
				starlark.String(err.Error()),
			}), nil
	}

	return starlark.Tuple([]starlark.Value{ret, starlark.None}), nil
}

func fail(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("error: got %d arguments, want 1", len(args))
	}

	buf := new(strings.Builder)
	stk := thread.CallStack()
	stk.Pop()

	fmt.Fprintf(buf, "%scustom error: ", stk)
	if s, ok := starlark.AsString(args[0]); ok {
		buf.WriteString(s)
	} else {
		buf.WriteString(args[0].String())
	}

	return nil, fmt.Errorf("%s", buf.String())
}

func assert(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("error: got %d arguments, want 2", len(args))
	}

	if args[0].Truth() {
		return starlark.None, nil
	}

	if len(args) < 2 {
		return nil, fmt.Errorf("assertion failed")
	}

	buf := new(strings.Builder)
	stk := thread.CallStack()
	stk.Pop()

	fmt.Fprintf(buf, "%scustom error: ", stk)
	if s, ok := starlark.AsString(args[1]); ok {
		buf.WriteString(s)
	} else {
		buf.WriteString(args[0].String())
	}

	return nil, fmt.Errorf("%s", buf.String())
}
