package values

import (
	"errors"
	"fmt"
	"time"

	starlarktime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (vctx *Context) ToStarlarkValue(v sdktypes.Value) (starlark.Value, error) {
	switch vv := v.Concrete().(type) {
	case sdktypes.FunctionValue:
		return vctx.functionToStarlark(v)
	case sdktypes.NothingValue:
		return starlark.None, nil
	case sdktypes.BooleanValue:
		return starlark.Bool(vv.Value()), nil
	case sdktypes.StringValue:
		return starlark.String(vv.Value()), nil
	case sdktypes.IntegerValue:
		return starlark.MakeInt64(vv.Value()), nil
	case sdktypes.BytesValue:
		return starlark.Bytes(vv.Value()), nil
	case sdktypes.FloatValue:
		return starlark.Float(vv.Value()), nil
	case sdktypes.SymbolValue:
		return newSymbol(vv.Symbol().String()), nil
	case sdktypes.DurationValue:
		return starlarktime.Duration(vv.Value()), nil
	case sdktypes.TimeValue:
		return starlarktime.Time(vv.Value()), nil
	case sdktypes.ListValue:
		vs, err := kittehs.TransformError(vv.Values(), vctx.ToStarlarkValue)
		if err != nil {
			return nil, fmt.Errorf("list: %w", err)
		}
		return starlark.NewList(vs), nil
	case sdktypes.SetValue:
		vs := vv.Values()
		set := starlark.NewSet(len(vs))

		for i, v := range vs {
			vv, err := vctx.ToStarlarkValue(v)
			if err != nil {
				return nil, fmt.Errorf("convert list.%d: %w", i, err)
			}

			if err := set.Insert(vv); err != nil {
				return nil, fmt.Errorf("insert list.%d: %w", i, err)
			}
		}
		return set, nil
	case sdktypes.DictValue:
		items := vv.Items()
		d := starlark.NewDict(len(items))
		for _, kv := range items {
			k, err := vctx.ToStarlarkValue(kv.K)
			if err != nil {
				return nil, fmt.Errorf("convert dict key: %w", err)
			}

			v, err := vctx.ToStarlarkValue(kv.V)
			if err != nil {
				return nil, fmt.Errorf("convert dict value: %w", err)
			}

			if err := d.SetKey(k, v); err != nil {
				return nil, fmt.Errorf("dict set: %w", err)
			}
		}
		return d, nil
	case sdktypes.StructValue:
		ctor, fields := vv.Ctor(), vv.Fields()

		sctor, err := vctx.ToStarlarkValue(ctor)
		if err != nil {
			return nil, fmt.Errorf("struct ctor: %w", err)
		}

		sfields, err := kittehs.TransformMapValuesError(fields, vctx.ToStarlarkValue)
		if err != nil {
			return nil, fmt.Errorf("struct fields: %w", err)
		}

		return starlarkstruct.FromStringDict(sctor, sfields), nil
	case sdktypes.ModuleValue:
		name, members := vv.Name(), vv.Members()

		smembers, err := kittehs.TransformMapValuesError(members, vctx.ToStarlarkValue)
		if err != nil {
			return nil, fmt.Errorf("struct fields: %w", err)
		}

		return &starlarkstruct.Module{
			Name:    name.String(),
			Members: smembers,
		}, nil
	default:
		return nil, sdkerrors.NewInvalidArgumentError("unrecognized type: %T", v)
	}
}

func (vctx *Context) FromStarlarkValue(v starlark.Value) (sdktypes.Value, error) {
	switch v := v.(type) {
	case *starlark.Builtin:
		return vctx.fromStarlarkBuiltin(v)
	case *starlark.Function:
		return vctx.fromStarlarkFunction(v)
	case starlark.NoneType:
		return sdktypes.Nothing, nil
	case starlark.Bool:
		return sdktypes.NewBooleanValue(bool(v)), nil
	case starlark.String:
		return sdktypes.NewStringValue(string(v)), nil
	case starlark.Int:
		i64, ok := v.Int64()
		if !ok {
			// TODO(ENG-61): support big int.
			return sdktypes.InvalidValue, errors.New("convert from starlark int")
		}
		return sdktypes.NewIntegerValue(i64), nil
	case starlark.Bytes:
		// TODO: Not sure that starlark's assumption that string and bytes are the same in go.
		return sdktypes.NewBytesValue([]byte(v)), nil
	case starlark.Float:
		return sdktypes.NewFloatValue(v), nil
	case *Symbol:
		s, err := sdktypes.ParseSymbol(string(*v))
		if err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("invalid Starlark symbol %q: %w", v.String(), err)
		}
		return sdktypes.NewSymbolValue(s), nil
	case starlarktime.Duration:
		return sdktypes.NewDurationValue(time.Duration(v)), nil
	case starlarktime.Time:
		return sdktypes.NewTimeValue(time.Time(v)), nil
	case *starlark.List, starlark.Tuple:
		return vctx.fromSequence(v.(starlark.Sequence), kittehs.Must11(sdktypes.NewListValue))
	case *starlark.Set:
		return vctx.fromSequence(v, kittehs.Must11(sdktypes.NewSetValue))
	case *starlark.Dict:
		ks := v.Keys()
		items := make([]sdktypes.DictItem, len(ks))
		for i, k := range ks {
			kv, err := vctx.FromStarlarkValue(k)
			if err != nil {
				return sdktypes.InvalidValue, fmt.Errorf("key conversion: %w", err)
			}

			v, found, err := v.Get(k)
			if err != nil {
				return sdktypes.InvalidValue, fmt.Errorf("dict value get: %w", err)
			} else if !found {
				return sdktypes.InvalidValue, errors.New("dict value missing")
			}

			vv, err := vctx.FromStarlarkValue(v)
			if err != nil {
				return sdktypes.InvalidValue, fmt.Errorf("value conversion: %w", err)
			}

			items[i] = sdktypes.DictItem{K: kv, V: vv}
		}

		return sdktypes.NewDictValue(items)
	case *starlarkstruct.Module:
		ms, err := kittehs.TransformMapValuesError(v.Members, vctx.FromStarlarkValue)
		if err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("module members: %w", err)
		}
		sym, err := sdktypes.ParseSymbol(v.Name)
		if err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("module name: %w", err)
		}
		return sdktypes.NewModuleValue(sym, ms)
	case *starlarkstruct.Struct:
		d := make(starlark.StringDict)
		v.ToStringDict(d)
		ms, err := kittehs.TransformMapValuesError(d, vctx.FromStarlarkValue)
		if err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("struct fields: %w", err)
		}
		ctor, err := vctx.FromStarlarkValue(v.Constructor())
		if err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("struct ctor: %w", err)
		}
		return sdktypes.NewStructValue(ctor, ms)
	default:
		return sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("unrecognized type: %T", v)
	}
}

func (vctx *Context) fromSequence(seq starlark.Sequence, f func([]sdktypes.Value) sdktypes.Value) (sdktypes.Value, error) {
	vs := make([]sdktypes.Value, seq.Len())

	it := seq.Iterate()
	defer it.Done()

	var sv starlark.Value
	for i := 0; it.Next(&sv); i++ {
		var err error
		if vs[i], err = vctx.FromStarlarkValue(sv); err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("list: %w", err)
		}
	}

	return sdktypes.NewListValue(vs)
}
