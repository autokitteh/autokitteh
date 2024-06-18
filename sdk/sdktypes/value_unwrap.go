package sdktypes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

func UnwrapValue(v Value) (any, error) { return DefaultValueWrapper.Unwrap(v) }

func (w ValueWrapper) UnwrapMap(m map[string]Value) (map[string]any, error) {
	return kittehs.TransformMapValuesError(m, w.Unwrap)
}

// Unwraps a value, converting it to a native go type.
func (w ValueWrapper) Unwrap(v Value) (any, error) {
	if w.Preunwrap != nil {
		var err error
		if v, err = w.Preunwrap(v); err != nil {
			return nil, err
		}
		if !v.IsValid() {
			return nil, nil
		}
	}

	return w.unwrap(v)
}

func (w *ValueWrapper) unwrap(v Value) (any, error) {
	if !v.IsValid() {
		return nil, nil
	}

	if v.IsFunction() && w.UnwrapFunction != nil {
		return w.UnwrapFunction(v)
	}

	switch v := v.Concrete().(type) {
	case NothingValue:
		return nil, nil
	case FunctionValue:
		return nil, errors.New("function values are not supported")
	case StringValue:
		return v.Value(), nil
	case IntegerValue:
		return v.Value(), nil
	case FloatValue:
		return v.Value(), nil
	case BooleanValue:
		return v.Value(), nil
	case BytesValue:
		return v.Value(), nil
	case TimeValue:
		return v.Value(), nil
	case DurationValue:
		d := v.Value()
		if w.RawDuration {
			return d, nil
		}
		return d, nil
	case SymbolValue:
		return v.Symbol().String(), nil
	case ListValue:
		return kittehs.TransformError(v.Values(), w.Unwrap)
	case SetValue:
		return kittehs.TransformError(v.Values(), w.Unwrap)
	case StructValue:
		if w.UnwrapStructsAsJSON {
			// nop - will marshal to JSON below.
			break
		}

		ctor, fields := v.Ctor(), v.Fields()
		m := make(map[string]any, len(fields)+1)
		var err error
		if m[ctorFieldName], err = w.Unwrap(ctor); err != nil {
			return nil, fmt.Errorf("ctor: %w", err)
		}
		for n, v := range fields {
			if m[n], err = w.Unwrap(v); err != nil {
				return nil, fmt.Errorf("field %q: %w", n, err)
			}
		}
		return m, nil

	case DictValue:
		d := v.Items()
		if w.SafeForJSON {
			if len(d) == 0 {
				return toJSONMap[string](d, w)
			}

			// Not all types are currently supported
			// If a new type is required, it should be added later
			k := d[0].K

			if k.IsString() {
				return toJSONMap[string](d, w)
			}

			if k.IsInteger() {
				return toJSONMap[int64](d, w)
			}

			if k.IsBoolean() {
				return toJSONMap[bool](d, w)
			}

			return nil, errors.New("unsupported dict key")
		}

		vs := make(map[any]any, len(d))
		for _, kv := range d {
			k, err := w.Unwrap(kv.K)
			if err != nil {
				return nil, err
			}

			v, err := w.Unwrap(kv.V)
			if err != nil {
				return nil, err
			}

			vs[k] = v
		}
		return vs, nil
	}

	if w.UnwrapUnknown != nil {
		return w.UnwrapUnknown(v)
	}

	return json.Marshal(v)
}

func toJSONMap[K comparable](items []DictItem, w *ValueWrapper) (map[K]any, error) {
	vs := make(map[K]any, len(items))

	for _, kv := range items {
		k, err := w.Unwrap(kv.K)
		if err != nil {
			return nil, err
		}

		if vs[k.(K)], err = w.Unwrap(kv.V); err != nil {
			return nil, fmt.Errorf("%q: %w", kv.K.String(), err)
		}
	}
	return vs, nil
}

func (w ValueWrapper) UnwrapInto(dst any, v Value) error {
	return w.unwrapInto("", reflect.ValueOf(dst), v)
}

func UnwrapValueInto(dst any, v Value) error { return DefaultValueWrapper.UnwrapInto(dst, v) }

func (w ValueWrapper) unwrapInto(path string, dstv reflect.Value, v Value) error {
	if w.Preunwrap != nil {
		v, err := w.Preunwrap(v)
		if err != nil {
			return err
		}
		if !v.IsValid() {
			return nil
		}

		return w.unwrapInto(path, dstv, v)
	}

	// prefix is the string to prepend to error messages to indicate what errored.
	if path != "" {
		path = fmt.Sprintf("%s: ", path)
	}

	// deref ptrs if needed.
	if dstv.Kind() == reflect.Ptr {
		return w.unwrapPtrInto(path, dstv, v)
	}

	if !dstv.CanSet() {
		return errors.New("dst is not settable")
	}

	if _, ok := dstv.Interface().(Value); ok {
		// Target is a Value, just assign it directly.
		dstv.Set(reflect.ValueOf(v))
		return nil
	}

	if IsConcreteValue(dstv.Interface()) {
		// Target is a concrete value which contains the actual value inside.
		// extract the actual value and assign it.

		vv := reflect.ValueOf(v.Concrete())

		if !vv.Type().AssignableTo(dstv.Type()) {
			return errors.New("cannot assign to non-value object")
		}

		dstv.Set(vv)
		return nil
	}

	if handled, err := w.unwrapContainerInto(path, dstv, v); handled {
		return err
	}

	if handled, err := w.unwrapScalarInto(path, dstv, v); handled {
		return err
	}

	return fmt.Errorf("%scannot unwrap into %v", path, dstv.Type())
}

func (w ValueWrapper) unwrapPtrInto(path string, dstv reflect.Value, v Value) error {
	if v.IsNothing() && dstv.CanSet() {
		// Target is a ptr and settable, and we need to
		// set nothing, so just set to nil.
		dstv.SetZero()
		return nil
	}

	if dstv.IsNil() {
		// Need to allocate a new ptr, to make elem settable.
		ptr := reflect.New(dstv.Type().Elem())
		dstv.Set(ptr)
	}

	return w.unwrapInto(path, dstv.Elem(), v)
}

func (w ValueWrapper) unwrapIntoReader(path string, dstv reflect.Value, v Value) (bool, error) {
	var buf bytes.Buffer

	switch v := v.Concrete().(type) {
	case NothingValue:
	case StringValue:
		if _, err := buf.WriteString(v.Value()); err != nil {
			return false, err
		}
	case BytesValue:
		if _, err := buf.Write(v.Value()); err != nil {
			return false, err
		}
	default:
		return true, fmt.Errorf("%scannot convert into reader", path)
	}

	dstv.Set(reflect.ValueOf(&buf))

	return true, nil
}

func (w ValueWrapper) unwrapScalarInto(path string, dstv reflect.Value, v Value) (bool, error) {
	if v.IsNothing() {
		if dstv.Kind() == reflect.Ptr {
			dstv.Set(reflect.Zero(dstv.Type()))
			return true, nil
		}

		return true, fmt.Errorf("%scannot convert Nothing to target type", path)
	}

	if dstv.Type().Implements(reflect.TypeOf((*io.Reader)(nil)).Elem()) {
		if w.IgnoreReader {
			dstv.Set(reflect.Zero(dstv.Type()))
			return true, nil
		}

		return w.unwrapIntoReader(path, dstv, v)
	}

	switch dstv.Interface().(type) {
	case time.Duration:
		d, err := v.ToDuration()
		if err != nil {
			return true, fmt.Errorf("%scannot convert duration", path)
		}

		dstv.Set(reflect.ValueOf(d))
		return true, nil

	case time.Time:
		t, err := v.ToTime()
		if err != nil {
			return true, fmt.Errorf("%scannot convert time", path)
		}

		dstv.Set(reflect.ValueOf(t))
		return true, nil

	default:
		u, err := w.unwrap(v)
		if err != nil {
			return true, fmt.Errorf("unwrap: %w", err)
		}

		if sym, ok := u.(Symbol); ok {
			u = sym.String()
		}

		uv := reflect.ValueOf(u)

		if uv.Type().AssignableTo(dstv.Type()) {
			dstv.Set(uv)
			return true, nil
		}

		if uv.Type().ConvertibleTo(dstv.Type()) {
			v1 := uv.Convert(dstv.Type())
			dstv.Set(v1)
			return true, nil
		}

		return false, nil
	}
}

func (w ValueWrapper) unwrapContainerInto(path string, dstv reflect.Value, v Value) (bool, error) {
	// path to supply to recurrent unwrapInto to indicate new value name.
	pathf := func(f string, xs ...any) string {
		p := path
		if p != "" {
			p += "."
		}

		return p + fmt.Sprintf(f, xs...)
	}

	dstt := dstv.Type()
	dstk := dstt.Kind()

	switch v := v.Concrete().(type) {
	case ListValue, SetValue:
		var isSet bool
		var vs []Value
		if lv, ok := v.(ListValue); ok {
			vs = lv.Values()
		} else {
			vs = v.(SetValue).Values()
			isSet = true
		}

		switch dstk {
		case reflect.Slice:
			slicev := reflect.MakeSlice(dstt, len(vs), len(vs))

			for i, lv := range vs {
				if err := w.unwrapInto(pathf("%d", i), slicev.Index(i), lv); err != nil {
					return true, err
				}
			}

			dstv.Set(slicev)
			return true, nil

		case reflect.Array:
			if dstv.Len() != len(vs) {
				break
			}

			for i, lv := range vs {
				if err := w.unwrapInto(pathf("%d", i), dstv.Index(i), lv); err != nil {
					return true, err
				}
			}

			return true, nil

		// for sets -> map[V]bool
		case reflect.Map:
			if !isSet || dstt.Elem().Kind() != reflect.Bool {
				break
			}

			mv := reflect.MakeMap(dstt)

			for i, lv := range vs {
				kv := reflect.New(dstt.Key())

				if err := w.unwrapInto(pathf("%d", i), kv, lv); err != nil {
					return true, err
				}

				mv.SetMapIndex(kv.Elem(), reflect.ValueOf(true))
			}

			dstv.Set(mv)
			return true, nil
		}

	case DictValue, StructValue:
		var items []DictItem

		if dv, ok := v.(DictValue); ok {
			items = dv.Items()
		} else {
			fields := v.(StructValue).Fields()

			items = kittehs.TransformMapToList(fields, func(n string, v Value) DictItem {
				return DictItem{
					K: NewStringValue(n),
					V: v,
				}
			})
		}

		switch dstk {
		case reflect.Map:
			mv := reflect.MakeMapWithSize(dstt, len(items))

			for _, item := range items {
				keyv := reflect.New(dstt.Key())
				if err := w.unwrapInto(pathf("key:%v", item.K), keyv, item.K); err != nil {
					return true, err
				}

				valv := reflect.New(dstt.Elem())
				if err := w.unwrapInto(pathf("value:%v", item.K), valv, item.V); err != nil {
					return true, err
				}

				mv.SetMapIndex(keyv.Elem(), valv.Elem())
			}

			dstv.Set(mv)
			return true, nil

		case reflect.Struct:
			sv := reflect.New(dstt)

			for _, item := range items {
				if !item.K.IsString() {
					return true, fmt.Errorf("%scannot convert non-string-key dict into struct", pathf(""))
				}

				k := item.K.GetString().Value()
				fn := w.toStructCaser(k)

				fv := sv.Elem().FieldByName(fn)
				if fv.Kind() == reflect.Invalid {
					if k != fn {
						fv = sv.Elem().FieldByName(k)
						if fv.Kind() == reflect.Invalid {
							if w.UnwrapErrorOnNonexistentStructFields {
								return true, fmt.Errorf("%s field %q or %q does not exit", pathf(""), fn, k)
							}

							continue
						}
					}

					if fv.Kind() == reflect.Invalid {
						return true, fmt.Errorf("%s field %q does not exit", pathf(""), k)
					}
				}

				if err := w.unwrapInto(pathf("%s", fn), fv, item.V); err != nil {
					return true, err
				}
			}

			dstv.Set(sv.Elem())
			return true, nil
		}
	}

	return false, nil
}
