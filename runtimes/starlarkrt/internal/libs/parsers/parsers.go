package parsers

import (
	"errors"
	"strings"

	"go.starlark.net/starlark"

	"go.autokitteh.dev/autokitteh/runtimes/configrt/parsers"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/values"
)

const ModuleName = "parsers"

func LoadModule() (starlark.StringDict, error) {
	return starlark.StringDict(map[string]starlark.Value{
		"parse": starlark.NewBuiltin("parse", parse),
	}), nil
}

func parse(thread *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var text, ext string

	if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "text", &text, "ext", &ext); err != nil {
		return nil, err
	}

	parse := parsers.Parsers[ext]
	if parse == nil {
		return nil, errors.New("no such parser")
	}

	v, err := parse(strings.NewReader(text))
	if err != nil {
		return nil, err
	}

	return (&values.Context{}).ToStarlarkValue(v)
}
