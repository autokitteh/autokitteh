package starlarkutils

import (
	"fmt"
	"reflect"

	"github.com/iancoleman/strcase"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type fromOpts struct {
	prefix  string
	nameCvt func(string) string
	valCvt  func(starlark.Value) (starlark.Value, error)
}

func WithFromNameConverter(nameCvt func(string) string) func(*fromOpts) {
	return func(o *fromOpts) { o.nameCvt = nameCvt }
}

func WithFromPrefix(p string) func(*fromOpts) {
	return func(o *fromOpts) { o.prefix = p }
}

func WithValCvt(cvt func(starlark.Value) (starlark.Value, error)) func(*fromOpts) {
	return func(o *fromOpts) { o.valCvt = cvt }
}

func FromStarlark(v starlark.Value, dst interface{}, optFuncs ...func(*fromOpts)) error {
	return fromStarlark(v, reflect.ValueOf(dst), optFuncs...)
}

func fromStarlark(v starlark.Value, dst reflect.Value, optFuncs ...func(*fromOpts)) error {
	opts := fromOpts{
		nameCvt: strcase.ToSnake,
		valCvt:  func(v starlark.Value) (starlark.Value, error) { return v, nil },
	}

	for _, opt := range optFuncs {
		opt(&opts)
	}

	prefix := opts.prefix
	chprefix := func(p string) []func(*fromOpts) { return append(optFuncs, WithFromPrefix(p)) }

	v, err := opts.valCvt(v)
	if err != nil {
		return fmt.Errorf("%s: %w", prefix, err)
	}

	switch dst.Type().Kind() {
	case reflect.Ptr:
		if dst.CanSet() {
			if v.Type() == starlark.None.Type() {
				dst.Set(reflect.Zero(dst.Type()))
				return nil
			}

			if !dst.Elem().IsValid() {
				dst.Set(reflect.New(dst.Type().Elem()))
			}

			return fromStarlark(v, dst.Elem(), optFuncs...)
		} else {
			if v.Type() == starlark.None.Type() {
				return nil
			}

			return fromStarlark(v, dst.Elem(), optFuncs...)
		}
	case reflect.String:
		s, ok := v.(starlark.String)
		if !ok {
			return fmt.Errorf("expected string, found %q", v.Type())
		}
		dst.SetString(string(s))
		return nil

	case reflect.Bool:
		b, ok := v.(starlark.Bool)
		if !ok {
			return fmt.Errorf("expected bool, found %q", v.Type())
		}
		dst.SetBool(bool(b))
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, ok := v.(starlark.Int)
		if !ok {
			return fmt.Errorf("expected int, found %q", v.Type())
		}
		i64, _ := i.Int64()
		dst.SetInt(i64)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, ok := v.(starlark.Int)
		if !ok {
			return fmt.Errorf("expected int, found %q", v.Type())
		}
		i64, _ := i.Uint64()
		dst.SetUint(i64)
		return nil

	case reflect.Float32, reflect.Float64:
		f, ok := v.(starlark.Float)
		if !ok {
			return fmt.Errorf("expected float, found %q", v.Type())
		}
		dst.SetFloat(float64(f))
		return nil

	case reflect.Slice:
		if v.Type() == starlark.None.Type() {
			return nil
		}

		if !dst.CanSet() {
			return fmt.Errorf("%s: not settable", prefix)
		}

		l, ok := v.(*starlark.List)
		if !ok {
			return fmt.Errorf("expected list, found %q", v.Type())
		}

		sl := reflect.MakeSlice(dst.Type(), l.Len(), l.Len())

		for i := 0; i < l.Len(); i++ {
			prefix = fmt.Sprintf("%s[%d]", prefix, i)
			if err := fromStarlark(l.Index(i), sl.Index(i), chprefix(prefix)...); err != nil {
				return fmt.Errorf("%s: %w", prefix, err)
			}
		}

		dst.Set(sl)
		return nil

	case reflect.Array:
		if v.Type() == starlark.None.Type() {
			return nil
		}

		l, ok := v.(*starlark.List)
		if !ok {
			return fmt.Errorf("expected list, found %q", v.Type())
		}

		if dst.Len() != l.Len() {
			return fmt.Errorf("list size %d != array size %d", l.Len(), dst.Len())
		}

		for i := 0; i < l.Len(); i++ {
			prefix = fmt.Sprintf("%s[%d]", prefix, i)
			if err := fromStarlark(l.Index(i), dst.Index(i), chprefix(prefix)...); err != nil {
				return fmt.Errorf("%s: %w", prefix, err)
			}
		}

		return nil

	case reflect.Map:
		if v.Type() == starlark.None.Type() {
			return nil
		}

		d, ok := v.(*starlark.Dict)
		if !ok {
			return fmt.Errorf("expected dict, found %q", v.Type())
		}

		m := reflect.MakeMapWithSize(dst.Type(), d.Len())

		for _, k := range d.Keys() {
			prefix := fmt.Sprintf("%s[%q]", prefix, k.String())

			v, _, _ := d.Get(k)

			kv := reflect.New(m.Type().Key())

			if err := fromStarlark(k, kv, chprefix(prefix)...); err != nil {
				return fmt.Errorf("%q: %w", prefix, err)
			}

			vv := reflect.New(m.Type().Elem())

			if err := fromStarlark(v, vv, chprefix(prefix)...); err != nil {
				return fmt.Errorf("%q: %w", prefix, err)
			}

			m.SetMapIndex(kv.Elem(), vv.Elem())
		}

		dst.Set(m)
		return nil

	case reflect.Struct:
		if v.Type() == starlark.None.Type() {
			return nil
		}

		var get func(string) (starlark.Value, bool)

		if d, ok := v.(*starlark.Dict); ok {
			get = func(k string) (starlark.Value, bool) {
				v, found, err := d.Get(starlark.String(k))
				if v == nil || !found || err != nil {
					return nil, false
				}
				return v, true
			}
		} else {
			d, ok := v.(*starlarkstruct.Struct)
			if !ok {
				return fmt.Errorf("expected struct, found %q", v.Type())
			}

			get = func(s string) (starlark.Value, bool) {
				v, err := d.Attr(s)
				if v == nil || err != nil {
					return nil, false
				}

				return v, true
			}
		}

		fs := reflect.VisibleFields(dst.Type())
		for _, f := range fs {
			if !f.IsExported() {
				continue
			}

			prefix = fmt.Sprintf("%s.%s", prefix, f.Name)

			fv := dst.FieldByIndex(f.Index)

			n := opts.nameCvt(f.Name)
			df, ok := get(n)
			if !ok {
				continue
			}

			if err := fromStarlark(df, fv, chprefix(prefix)...); err != nil {
				return fmt.Errorf("%q: %w", prefix, err)
			}
		}

		return nil

	default:
		return fmt.Errorf("unhandled kind %q", dst.Type().Kind())
	}
}
