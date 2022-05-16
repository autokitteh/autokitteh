package starlarkutils

import (
	"fmt"
	"reflect"

	"github.com/iancoleman/strcase"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type toOpts struct {
	prefix  string
	nameCvt func(string) string
	valCvt  func(interface{}) (interface{}, error)
}

func WithToPrefix(p string) func(*toOpts) {
	return func(o *toOpts) { o.prefix = p }
}

func WithToNameConverter(nameCvt func(string) string) func(*toOpts) {
	return func(o *toOpts) { o.nameCvt = nameCvt }
}

func WithToValueConverter(cvt func(interface{}) (interface{}, error)) func(*toOpts) {
	return func(o *toOpts) { o.valCvt = cvt }
}

func ToStarlark(x interface{}, optFuncs ...func(*toOpts)) (starlark.Value, error) {
	opts := toOpts{
		nameCvt: strcase.ToSnake,
		valCvt:  func(v interface{}) (interface{}, error) { return v, nil },
	}

	for _, opt := range optFuncs {
		opt(&opts)
	}

	x, err := opts.valCvt(x)
	if err != nil {
		return nil, fmt.Errorf("%s: convert: %w", opts.prefix, err)
	}

	if x == nil {
		return starlark.None, nil
	}

	chprefix := func(p string) []func(*toOpts) { return append(optFuncs, WithToPrefix(p)) }

	v, t := reflect.ValueOf(x), reflect.TypeOf(x)

	switch t.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return starlark.None, nil
		}
		return ToStarlark(v.Elem().Interface(), optFuncs...)
	case reflect.String:
		return starlark.String(v.String()), nil
	case reflect.Bool:
		return starlark.Bool(v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return starlark.MakeInt64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return starlark.MakeUint64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return starlark.Float(v.Float()), nil
	case reflect.Slice, reflect.Array:
		elems := make([]starlark.Value, v.Len())
		for i := 0; i < v.Len(); i++ {
			var err error
			prefix := fmt.Sprintf("%s[%d]", opts.prefix, i)
			if elems[i], err = ToStarlark(v.Index(i).Interface(), chprefix(prefix)...); err != nil {
				return nil, fmt.Errorf("%s: %w", prefix, err)
			}
		}
		return starlark.NewList(elems), nil
	case reflect.Map:
		d := starlark.NewDict(v.Len())
		for _, k := range v.MapKeys() {
			prefix := fmt.Sprintf("%s[%q]", opts.prefix, k.String())
			vv, err := ToStarlark(v.MapIndex(k).Interface(), chprefix(prefix)...)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", prefix, err)
			}
			_ = d.SetKey(starlark.String(k.String()), vv)
		}
		return d, nil
	case reflect.Struct:
		vfs := reflect.VisibleFields(t)
		ms := make(map[string]starlark.Value, len(vfs))
		for _, vf := range vfs {
			if !vf.IsExported() {
				continue
			}

			prefix := fmt.Sprintf("%s.%s", opts.prefix, vf.Name)
			fv := v.FieldByIndex(vf.Index)

			var err error
			ms[opts.nameCvt(vf.Name)], err = ToStarlark(fv.Interface(), chprefix(prefix)...)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", prefix, err)
			}
		}
		return starlarkstruct.FromStringDict(Symbol(opts.nameCvt(t.Name())), ms), nil
	default:
		return nil, fmt.Errorf("%s: unhandled type %q", opts.prefix, t.Kind())
	}
}
