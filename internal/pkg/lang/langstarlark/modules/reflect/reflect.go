package reflect

import (
	"errors"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var Module = starlarkstruct.Module{
	Name: "reflect",
	Members: map[string]starlark.Value{
		"funcargs": starlark.NewBuiltin("funcargs", funcargs),
	},
}

func Load() map[string]starlark.Value { return Module.Members }

func funcargs(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var fn starlark.Value

	if err := starlark.UnpackArgs("funcargs", args, kwargs, "fn", &fn); err != nil {
		return nil, err
	}

	starfn, ok := fn.(*starlark.Function)
	if !ok {
		return nil, errors.New("argument is not a starlark function")
	}

	l := make([]starlark.Value, starfn.NumParams())
	for i := 0; i < starfn.NumParams(); i++ {
		p, _ := starfn.Param(i)
		l[i] = starlark.String(p)
	}

	return starlark.NewList(l), nil
}
