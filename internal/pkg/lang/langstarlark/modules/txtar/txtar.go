package txtar

import (
	"fmt"

	"golang.org/x/tools/txtar"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var Module = starlarkstruct.Module{
	Name: "txtar",
	Members: map[string]starlark.Value{
		"parse":  starlark.NewBuiltin("parse", parse),
		"format": starlark.NewBuiltin("format", format),
	},
}

func Load() map[string]starlark.Value { return Module.Members }

func format(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var (
		comment starlark.String
		fs      *starlark.Dict
	)

	if err := starlark.UnpackArgs("parse", args, kwargs, "comment", &comment, "files", &fs); err != nil {
		return nil, err
	}

	arch := txtar.Archive{
		Comment: []byte(comment),
		Files:   make([]txtar.File, fs.Len()),
	}

	for i, k := range fs.Keys() {
		ks, ok := k.(starlark.String)
		if !ok {
			return nil, fmt.Errorf("files must be a dictionary from string to string")
		}

		v, found, err := fs.Get(k)
		if err != nil {
			return nil, fmt.Errorf("get file %q: %w", k, err)
		} else if !found {
			return nil, fmt.Errorf("get file %q: not found", k)
		}

		vs, ok := v.(starlark.String)
		if !ok {
			return nil, fmt.Errorf("file %q: not a string", k)
		}

		arch.Files[i] = txtar.File{
			Name: string(ks),
			Data: []byte(vs),
		}
	}

	return starlark.String(txtar.Format(&arch)), nil
}

func parse(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var text starlark.String

	if err := starlark.UnpackArgs("parse", args, kwargs, "text", &text); err != nil {
		return nil, err
	}

	arch := txtar.Parse([]byte(text))

	fs := starlark.NewDict(len(arch.Files))

	for _, f := range arch.Files {
		if err := fs.SetKey(starlark.String(f.Name), starlark.String(f.Data)); err != nil {
			return nil, fmt.Errorf("set %q: %w", f.Name, err)
		}
	}

	return starlark.Tuple{starlark.String(arch.Comment), fs}, nil
}
