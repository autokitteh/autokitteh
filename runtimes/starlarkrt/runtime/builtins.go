package runtime

import (
	"errors"
	"fmt"

	"go.starlark.net/starlark"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var builtins = map[string]func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error){
	"catch": func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		// TODO: optionally do not catch starlark script syntax errors, just actually failures.

		fn, ok := args[0].(starlark.Callable)
		if !ok {
			return nil, errors.New("first argument must be a function")
		}

		value, err := starlark.Call(thread, fn, args[1:], kwargs)
		if err != nil {
			return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
		}

		return starlark.Tuple{value, starlark.None}, nil
	},
	"fail": func(thread *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var msg string
		if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "msg", &msg); err != nil {
			return nil, err
		}
		return nil, errors.New(msg)
	},
	"globals": func(thread *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if err := starlark.UnpackArgs(bi.Name(), args, kwargs); err != nil {
			return nil, err
		}

		globals, ok := thread.Local(globalsTLSKey).(starlark.StringDict)
		if !ok {
			return nil, fmt.Errorf("globals are not set")
		}

		d := starlark.NewDict(len(globals))
		for k, v := range globals {
			kittehs.Must0(d.SetKey(starlark.String(k), v))
		}

		return d, nil
	},
}
