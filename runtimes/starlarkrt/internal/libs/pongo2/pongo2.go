package pongo2

import (
	"errors"
	"fmt"

	"github.com/flosch/pongo2/v6"
	"github.com/psanford/memfs"
	"go.starlark.net/starlark"

	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/values"
)

const ModuleName = "pongo2"

func LoadModule() (starlark.StringDict, error) {
	return starlark.StringDict(map[string]starlark.Value{
		"render_string": starlark.NewBuiltin("render_string", renderString),
		"render_dict":   starlark.NewBuiltin("render_dict", renderDict),
	}), nil
}

func render(tpl *pongo2.Template, context *starlark.Dict) (starlark.Value, error) {
	pongoContext := make(map[string]any, context.Len())
	for _, kv := range context.Items() {
		k, ok := kv[0].(starlark.String)
		if !ok {
			return nil, errors.New("context keys must be all strings")
		}

		v, err := values.Unwrap(kv[1])
		if err != nil {
			return nil, fmt.Errorf("context key %q: %w", k, err)
		}

		pongoContext[string(k)] = v
	}

	out, err := tpl.Execute(pongoContext)
	if err != nil {
		return nil, fmt.Errorf("execute: %w", err)
	}

	return starlark.String(out), nil
}

func renderString(thread *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		text    string
		context *starlark.Dict
	)

	if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "text", &text, "context", &context); err != nil {
		return nil, err
	}

	set := pongo2.NewSet("pongo2", pongo2.NewFSLoader(memfs.New()))

	tpl, err := set.FromString(text)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return render(tpl, context)
}

func renderDict(thread *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		root    string
		data    *starlark.Dict
		context *starlark.Dict
	)

	if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "data", &data, "root", &root, "context", &context); err != nil {
		return nil, err
	}

	fs := memfs.New()

	for _, kv := range data.Items() {
		path, ok := kv[0].(starlark.String)
		if !ok {
			return nil, errors.New("data must have string keys")
		}

		data, ok := kv[1].(starlark.String)
		if !ok {
			dataBytes, ok := kv[1].(starlark.Bytes)
			if !ok {
				return nil, errors.New("data must have either bytes or string values")
			}

			// starlark.Bytes is a string, so this works.
			data = starlark.String(dataBytes)
		}

		if err := fs.WriteFile(string(path), []byte(data), 0o600); err != nil {
			return nil, fmt.Errorf("write %q: %w", path, err)
		}
	}

	set := pongo2.NewSet("pongo2", pongo2.NewFSLoader(fs))

	tpl, err := set.FromFile(root)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return render(tpl, context)
}
