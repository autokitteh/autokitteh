package ak

import (
	"errors"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/tls"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/values"
)

var store = &starlarkstruct.Module{
	Name: "store",
	Members: starlark.StringDict{
		"get":       starlark.NewBuiltin("get", get),
		"list_keys": starlark.NewBuiltin("list_keys", listKeys),
		"mutate":    starlark.NewBuiltin("mutate", mutate),
	},
}

func get(th *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key string

	if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "key", &key); err != nil {
		return nil, err
	}

	return mutate(th, bi, starlark.Tuple{starlark.String(key), starlark.String("get")}, nil)
}

func listKeys(th *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs(bi.Name(), args, kwargs); err != nil {
		return nil, err
	}

	tls := tls.Get(th)

	list := tls.Callbacks.ListStoreValues
	if list == nil {
		return nil, errors.New("store.list not implemented")
	}

	keys, err := list(tls.GoCtx, tls.RunID)
	if err != nil {
		return nil, err
	}

	return starlark.NewList(kittehs.Transform(keys, func(s string) starlark.Value { return starlark.String(s) })), nil
}

func mutate(th *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		key, op     string
		operandsArg starlark.Value
	)

	if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "key", &key, "op", &op, "operands?", &operandsArg); err != nil {
		return nil, err
	}

	var operands []starlark.Value

	if operandsArg != nil && operandsArg.Truth() {
		iter := starlark.Iterate(operandsArg)
		if iter == nil {
			return nil, errors.New("operands must be iterable")
		}

		var v starlark.Value
		for iter.Next(&v) {
			operands = append(operands, v)
		}

		iter.Done()
	}

	values := values.FromTLS(th)

	akOperands, err := kittehs.TransformError(operands, values.FromStarlarkValue)
	if err != nil {
		return nil, err
	}

	tls := tls.Get(th)

	mutate := tls.Callbacks.MutateStoreValue
	if mutate == nil {
		return nil, errors.New("store.mutate not implemented")
	}

	v, err := mutate(tls.GoCtx, tls.RunID, key, op, akOperands...)
	if err != nil {
		return nil, err
	}

	return values.ToStarlarkValue(v)
}
