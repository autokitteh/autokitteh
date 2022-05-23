package langstarlark

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	starlarktime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"go.autokitteh.dev/sdk/api/apivalues"
	"github.com/autokitteh/idgen"
	"github.com/autokitteh/starlarkutils"
)

var (
	ErrUnknownType = errors.New("(starlark) unknown type")
)

type values struct {
	lng   *langstarlark
	env   *lang.RunEnv
	funcs map[string]*starlark.Function // TODO: eviction.
}

func newValues(lng *langstarlark) *values {
	return &values{
		lng:   lng,
		funcs: make(map[string]*starlark.Function),
	}
}

func (v *values) WithEnv(env *lang.RunEnv) *values {
	vv := *v
	vv.env = env
	return &vv
}

func NewNaiveValues() *values { return &values{} }

var NilValues *values = nil // must always be nil

func (lv *values) FromStringDict(d starlark.StringDict, override, other func(starlark.Value) (*apivalues.Value, error)) (m map[string]*apivalues.Value, err error) {
	m = make(map[string]*apivalues.Value, len(d))
	for k, v := range d {
		if m[k], err = lv.FromStarlarkValue(v, override, other); err != nil {
			return nil, fmt.Errorf("key %q: %w", k, err)
		}
	}
	return
}

func (lv *values) ToStringDict(m map[string]*apivalues.Value) (d starlark.StringDict, err error) {
	d = make(starlark.StringDict, len(m))
	for k, v := range m {
		if d[k], err = lv.ToStarlarkValue(v); err != nil {
			return nil, fmt.Errorf("key %q: %w", k, err)
		}
	}
	return
}

func (lv *values) FromStarlarkValue(v starlark.Value, override, other func(starlark.Value) (*apivalues.Value, error)) (*apivalues.Value, error) {
	defaultOther := func(v starlark.Value) (*apivalues.Value, error) {
		return nil, fmt.Errorf("(from) %w: %v", ErrUnknownType, reflect.TypeOf(v))
	}

	if override != nil {
		if v1, err := override(v); err != nil {
			return nil, err
		} else if v1 != nil {
			return v1, nil
		}
	}

	switch vv := v.(type) {
	case starlark.NoneType:
		return apivalues.None, nil
	case starlark.String:
		return apivalues.String(string(vv)), nil
	case starlarkutils.Symbol:
		return apivalues.Symbol(string(vv)), nil
	case starlark.Int:
		i64, ok := vv.Int64()
		if !ok {
			return nil, fmt.Errorf("int64 get not ok")
		}
		return apivalues.NewValue(apivalues.IntegerValue(i64))
	case starlark.Bool:
		return apivalues.NewValue(apivalues.BooleanValue(vv))
	case starlark.Float:
		return apivalues.NewValue(apivalues.FloatValue(vv))
	case starlark.Bytes:
		return apivalues.NewValue(apivalues.BytesValue(vv))
	case starlark.Tuple:
		vs, err := lv.froms([]starlark.Value(vv), override, other)
		if err != nil {
			return nil, err
		}
		return apivalues.NewValue(apivalues.ListValue(vs))
	case starlarktime.Duration:
		return apivalues.NewValue(apivalues.DurationValue(vv))
	case starlarktime.Time:
		return apivalues.NewValue(apivalues.TimeValue(vv))
	case *starlark.List:
		vs := make([]*apivalues.Value, vv.Len())
		for i := 0; i < vv.Len(); i++ {
			var err error
			if vs[i], err = lv.FromStarlarkValue(vv.Index(i), override, other); err != nil {
				return nil, fmt.Errorf("item %d: %w", i, err)
			}
		}
		return apivalues.NewValue(apivalues.ListValue(vs))
	case *starlark.Set:
		vs := make([]starlark.Value, vv.Len())
		iter := vv.Iterate()
		for i := 0; iter.Next(&vs[i]); i++ {
			// nop
		}
		vvs, err := lv.froms(vs, override, other)
		if err != nil {
			return nil, err
		}
		return apivalues.NewValue(apivalues.SetValue(vvs))
	case *starlark.Dict:
		vs := make([]*apivalues.DictItem, vv.Len())
		for i, k := range vv.Keys() {
			var err error

			vs[i] = &apivalues.DictItem{}

			vs[i].K, err = lv.FromStarlarkValue(k, override, other)
			if err != nil {
				return nil, fmt.Errorf("key %v: %w", k, err)
			}

			v, found, err := vv.Get(k)
			if !found {
				panic("item not found, but it should have")
			} else if err != nil {
				return nil, fmt.Errorf("value for key %v: %w", k, err)
			}

			vs[i].V, err = lv.FromStarlarkValue(v, override, other)
			if err != nil {
				return nil, fmt.Errorf("key %v: value %v: %w", k, v, err)
			}
		}
		return apivalues.NewValue(apivalues.DictValue(vs))
	case *starlarkstruct.Struct:
		ctor, err := lv.FromStarlarkValue(vv.Constructor(), override, other)
		if err != nil {
			return nil, fmt.Errorf("ctor: %w", err)
		}

		d := make(starlark.StringDict)
		vv.ToStringDict(d)

		fs, err := lv.FromStringDict(d, override, other)
		if err != nil {
			return nil, fmt.Errorf("fields: %w", err)
		}

		return apivalues.NewValue(apivalues.StructValue{Ctor: ctor, Fields: fs})
	case *starlarkstruct.Module:
		ms, err := lv.FromStringDict(vv.Members, override, other)
		if err != nil {
			return nil, fmt.Errorf("members: %w", err)
		}

		return apivalues.NewValue(apivalues.ModuleValue{Name: vv.Name, Members: ms})
	case *starlark.Function:
		return lv.storeFunc(vv)
	case *starlark.Builtin:
		parts := strings.SplitN(vv.Name(), "|", 4)
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid builtin name")
		}

		fparts := strings.Split(parts[3], ",")
		flags := make(map[string]bool, len(fparts))
		for _, fp := range fparts {
			flags[fp] = true
		}

		return apivalues.NewValue(apivalues.CallValue{ID: parts[0], Issuer: parts[1], Name: parts[2], Flags: flags})
	default:
		if other != nil {
			if translated, err := other(v); err != nil {
				return nil, err
			} else if translated != nil {
				return translated, nil
			}
		}

		return defaultOther(v)
	}
}

func (lv *values) froms(ins []starlark.Value, override, other func(starlark.Value) (*apivalues.Value, error)) (outs []*apivalues.Value, err error) {
	outs = make([]*apivalues.Value, len(ins))
	for i, in := range ins {
		if outs[i], err = lv.FromStarlarkValue(in, override, other); err != nil {
			return nil, fmt.Errorf("item %d: %w", i, err)
		}
	}
	return
}

func (lv *values) ToStarlarkValue(v *apivalues.Value) (starlark.Value, error) {
	switch vv := v.Get().(type) {
	case apivalues.NoneValue:
		return starlark.None, nil
	case apivalues.StringValue:
		return starlark.String(vv), nil
	case apivalues.SymbolValue:
		return starlarkutils.Symbol(string(vv)), nil
	case apivalues.IntegerValue:
		return starlark.MakeInt(int(vv)), nil
	case apivalues.BooleanValue:
		return starlark.Bool(vv), nil
	case apivalues.FloatValue:
		return starlark.Float(float64(vv)), nil
	case apivalues.TimeValue:
		return starlarktime.Time(vv), nil
	case apivalues.DurationValue:
		return starlarktime.Duration(vv), nil
	case apivalues.BytesValue:
		return starlark.Bytes(string(vv)), nil
	case apivalues.ListValue:
		vs, err := lv.tos([]*apivalues.Value(vv))
		if err != nil {
			return nil, err
		}

		return starlark.NewList(vs), nil
	case apivalues.SetValue:
		vs, err := lv.tos([]*apivalues.Value(vv))
		if err != nil {
			return nil, err
		}

		set := starlark.NewSet(len(vs))
		for _, v := range vs {
			if err := set.Insert(v); err != nil {
				return nil, fmt.Errorf("set %v: %w", v, err)
			}
		}

		return set, nil
	case apivalues.DictValue:
		d := starlark.NewDict(len(vv))

		for _, kv := range vv {
			k, err := lv.ToStarlarkValue(kv.K)
			if err != nil {
				return nil, fmt.Errorf("key: %w", err)
			}

			v, err := lv.ToStarlarkValue(kv.V)
			if err != nil {
				return nil, fmt.Errorf("value: %w", err)
			}

			if err := d.SetKey(k, v); err != nil {
				return nil, fmt.Errorf("set %v=%v: %w", k, v, err)
			}
		}

		return d, nil
	case apivalues.CallValue:
		flags := make([]string, 0, len(vv.Flags))
		for k, v := range vv.Flags {
			if v {
				flags = append(flags, k)
			}
		}

		return starlark.NewBuiltin(
			fmt.Sprintf("%s|%s|%s|%s", vv.ID, vv.Issuer, vv.Name, strings.Join(flags, ",")),
			func(th *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kvs []starlark.Tuple) (starlark.Value, error) {
				rv, err := lv.builtin(th, vv, args, kvs)
				return rv, err
			},
		), nil
	case apivalues.StructValue:
		ctor, err := lv.ToStarlarkValue(vv.Ctor)
		if err != nil {
			return nil, fmt.Errorf("ctor: %w", err)
		}

		fs, err := lv.ToStringDict(vv.Fields)
		if err != nil {
			return nil, fmt.Errorf("fields: %w", err)
		}

		return starlarkstruct.FromStringDict(ctor, fs), nil
	case apivalues.ModuleValue:
		ms, err := lv.ToStringDict(vv.Members)
		if err != nil {
			return nil, fmt.Errorf("members: %w", err)
		}

		return &starlarkstruct.Module{
			Name:    vv.Name,
			Members: ms,
		}, nil
	case apivalues.FunctionValue:
		return lv.retreiveFunc(vv)
	default:
		return nil, fmt.Errorf("(to) %w: %v", ErrUnknownType, reflect.TypeOf(v))
	}
}

func (lv *values) tos(ins []*apivalues.Value) (outs []starlark.Value, err error) {
	outs = make([]starlark.Value, len(ins))
	for i, in := range ins {
		if outs[i], err = lv.ToStarlarkValue(in); err != nil {
			return nil, fmt.Errorf("item %d: %w", i, err)
		}
	}
	return
}

func (lv *values) builtin(
	th *starlark.Thread,
	cv apivalues.CallValue,
	slargs starlark.Tuple,
	slkvs []starlark.Tuple,
) (starlark.Value, error) {
	if lv == nil || lv.env.Call == nil {
		return nil, fmt.Errorf("builtin calls not supported")
	}

	var err error

	args := make([]*apivalues.Value, len(slargs))
	for i, slarg := range slargs {
		if args[i], err = lv.FromStarlarkValue(slarg, nil, nil); err != nil {
			return nil, fmt.Errorf("pos arg %d: %w", i, err)
		}
	}

	kvs := make(map[string]*apivalues.Value, len(slkvs))
	for _, slkv := range slkvs {
		if len(slkv) != 2 {
			return nil, fmt.Errorf("invalid starlark kv args, not a pair (len=%d)", len(slkv))
		}

		slk, slv := slkv[0], slkv[1]

		slks, ok := slk.(starlark.String)
		if !ok {
			return nil, fmt.Errorf("invalid starlark kv args, key not a string (%v)", slk.Type())
		}

		v, err := lv.FromStarlarkValue(slv, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to decode pair starlark value for key %q: %w", slks, err)
		}

		kvs[strings.TrimPrefix(string(slks), "secret_")] = v
	}

	// The env from the TLS is used here as the calling env might not be the same
	// as the env that is used to generate the builtin. This happens when a builtin
	// that is generated by a dependency is called from the dependee. Same idea
	// as context.
	rv, err := getTLSEnv(th).Call(getTLSContext(th), apivalues.MustNewValue(cv), kvs, args, nil)
	if err != nil {
		return nil, err
	}

	return lv.ToStarlarkValue(rv)
}

func (lv *values) storeFunc(f *starlark.Function) (*apivalues.Value, error) {
	if lv == nil {
		return nil, fmt.Errorf("passing functions externally is not supported")
	}

	argsNames := make([]string, f.NumParams())
	for i := 0; i < f.NumParams(); i++ {
		argsNames[i], _ = f.Param(i)
	}

	v := apivalues.FunctionValue{
		Lang:   lv.lng.name,
		FuncID: idgen.New("F"),
		Scope:  lv.env.Scope,
		Signature: &apivalues.FunctionSignature{
			Name:          f.Name(),
			Doc:           f.Doc(),
			NumArgs:       uint32(f.NumParams()),
			NumKWOnlyArgs: uint32(f.NumKwonlyParams()),
			ArgsNames:     argsNames,
		},
	}

	lv.funcs[v.FuncID] = f

	return apivalues.NewValue(v)
}

func (lv *values) retreiveFunc(f apivalues.FunctionValue) (*starlark.Function, error) {
	if lv == nil || lv.lng == nil {
		return nil, fmt.Errorf("functions as values are not supported")
	}

	if f.Lang != lv.lng.name {
		// TODO: direct to correct lang (inter-lang call).
		return nil, fmt.Errorf("lang mismatch: %s != current %s", f.Lang, lv.lng.name)
	}

	if f.Scope != lv.env.Scope {
		return nil, fmt.Errorf("scope mismatch: %s != current %s", f.Scope, lv.env.Scope)
	}

	slf, ok := lv.funcs[f.FuncID]
	if !ok {
		return nil, fmt.Errorf("function not found: %s", f.FuncID)
	}

	return slf, nil
}
