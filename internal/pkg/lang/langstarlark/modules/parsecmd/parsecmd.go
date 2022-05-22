package parsecmd

import (
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"github.com/autokitteh/parsecmd"
	"github.com/autokitteh/starlarkutils"
)

var Module = starlarkstruct.Module{
	Name: "parsecmd",
	Members: map[string]starlark.Value{
		"parsecmd": starlark.NewBuiltin("parsecmd", slParseCmd),
	},
}

func Load() map[string]starlark.Value { return Module.Members }

func slParseCmd(
	_ *starlark.Thread,
	bi *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var text string

	if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "text", &text); err != nil {
		return nil, err
	}

	cmd, err := parsecmd.Parse(text)
	if err != nil {
		return nil, err
	}

	return starlarkutils.ToStarlark(cmd)
}
