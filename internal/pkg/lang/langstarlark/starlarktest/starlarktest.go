// Adapted from https://github.com/google/starlark-go/blob/master/starlarktest/starlarktest.go.
package starlarktest

import (
	"fmt"
	"strings"

	"go.starlark.net/starlark"
)

var FailBuiltin = starlark.NewBuiltin("fail", fail)
var CatchBuiltin = starlark.NewBuiltin("catch", catch)

// catch(f) evaluates f() and returns its evaluation error message
// if it failed or None if it succeeded.
func catch(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var fn starlark.Callable
	if err := starlark.UnpackArgs("catch", args, kwargs, "fn", &fn); err != nil {
		return nil, err
	}
	if _, err := starlark.Call(thread, fn, nil, nil); err != nil {
		return starlark.String(err.Error()), nil
	}
	return starlark.None, nil
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
