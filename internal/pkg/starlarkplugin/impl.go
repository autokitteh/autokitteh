package starlarkplugin

import (
	"context"
	"fmt"

	"go.starlark.net/starlark"

	"github.com/autokitteh/autokitteh/internal/pkg/lang/langstarlark"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/pluginimpl"
)

func mkcall(bi *starlark.Builtin) func(
	context.Context,
	string,
	[]*apivalues.Value,
	map[string]*apivalues.Value,
	pluginimpl.FuncToValueFunc,
) (*apivalues.Value, error) {
	return func(
		ctx context.Context,
		name string,
		args []*apivalues.Value,
		params map[string]*apivalues.Value,
		funcToValue pluginimpl.FuncToValueFunc,
	) (*apivalues.Value, error) {
		// TODO: Cancellation (will it even work for builtins?)
		slargs := make([]starlark.Value, len(args))
		for i, a := range args {
			vv, err := langstarlark.NilValues.ToStarlarkValue(a)
			if err != nil {
				return nil, fmt.Errorf("param %d: %w", i, err)
			}

			slargs[i] = vv
		}

		kwargs := make([]starlark.Tuple, 0, len(params))
		for k, p := range params {
			kv := starlark.String(k)
			vv, err := langstarlark.NilValues.ToStarlarkValue(p)
			if err != nil {
				return nil, fmt.Errorf("param %q: %w", name, err)
			}

			kwargs = append(kwargs, starlark.Tuple([]starlark.Value{kv, vv}))
		}

		thr := starlark.Thread{Name: fmt.Sprintf("call:%s", name)}

		v, err := starlark.Call(&thr, bi, starlark.Tuple(slargs), kwargs)
		if err != nil {
			return nil, err
		}

		return langstarlark.NilValues.FromStarlarkValue(
			v,
			func(v starlark.Value) (*apivalues.Value, error) {
				bi, ok := v.(*starlark.Builtin)
				if !ok {
					return nil, nil
				}

				return funcToValue(bi.Name(), mkcall(bi)), nil
			},
			nil,
		)
	}
}

func Plugin(
	doc string,
	members starlark.StringDict,
) *pluginimpl.Plugin {
	plmembers := make(map[string]*pluginimpl.PluginMember, len(members))

	for name, v := range members {
		if bi, ok := v.(*starlark.Builtin); ok {
			// TODO: builtin doc somehow.
			plmembers[name] = pluginimpl.NewMethodMember("?", mkcall(bi))
		} else if v, err := langstarlark.NilValues.FromStarlarkValue(v, nil, nil); err != nil {
			panic(fmt.Errorf("%s: %w", name, err))
		} else {
			plmembers[name] = pluginimpl.NewValueMember("", v)
		}
	}

	return &pluginimpl.Plugin{
		Doc:     doc,
		Members: plmembers,
	}
}
