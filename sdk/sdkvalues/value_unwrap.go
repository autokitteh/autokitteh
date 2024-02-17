package sdkvalues

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func Unwrap(v sdktypes.Value) (any, error) { return DefaultValueWrapper.Unwrap(v) }

// Unwraps a value, converting it to a native go type.
func (w ValueWrapper) Unwrap(v sdktypes.Value) (any, error) {
	if w.Preunwrap != nil {
		var err error
		if v, err = w.Preunwrap(v); err != nil {
			return nil, err
		}
	}

	return w.unwrap(v)
}

func (w *ValueWrapper) unwrap(v sdktypes.Value) (any, error) {
	if v == nil {
		return nil, nil
	}

	switch sdktypes.GetValue(v).(type) {
	case sdktypes.NothingValue:
		return struct{}{}, nil
	case sdktypes.FunctionValue:
		return nil, errors.New("function values are not supported")
	case sdktypes.StringValue:
		return sdktypes.GetStringValue(v), nil
	case sdktypes.IntegerValue:
		return sdktypes.GetIntegerValue(v), nil
	case sdktypes.FloatValue:
		return sdktypes.GetFloatValue(v), nil
	case sdktypes.BooleanValue:
		return sdktypes.GetBooleanValue(v), nil
	case sdktypes.BytesValue:
		return sdktypes.GetBytesValue(v), nil
	case sdktypes.TimeValue:
		return sdktypes.GetTimeValue(v), nil
	case sdktypes.DurationValue:
		d := sdktypes.GetDurationValue(v)
		if w.RawDuration {
			return d, nil
		}
		return d, nil
	case sdktypes.SymbolValue:
		return sdktypes.GetSymbolValue(v), nil
	case sdktypes.ListValue:
		return kittehs.TransformError(sdktypes.GetListValue(v), w.Unwrap)
	case sdktypes.SetValue:
		return kittehs.TransformError(sdktypes.GetSetValue(v), w.Unwrap)
	case sdktypes.StructValue:
		if w.UnwrapStructsAsJSON {
			// nop - will marshal to JSON below.
			break
		}

		ctor, fields := sdktypes.GetStructValue(v)
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

	case sdktypes.DictValue:
		d := sdktypes.GetDictValue(v)
		if w.SafeForJSON {
			if len(d) == 0 {
				return toJSONMap[string](d, w)
			}

			// Not all types are currently supported
			// If a new type is required, it should be added later
			if sdktypes.IsStringValue(d[0].K) {
				return toJSONMap[string](d, w)
			}

			if sdktypes.IsIntegerValue(d[0].K) {
				return toJSONMap[int64](d, w)
			}

			if sdktypes.IsBooleanValue(d[0].K) {
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

func toJSONMap[K comparable](items []*sdktypes.DictValueItem, w *ValueWrapper) (map[K]any, error) {
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

func (w ValueWrapper) UnwrapInto(dst any, v sdktypes.Value) error {
	return w.unwrapInto("", reflect.ValueOf(dst), v)
}

func (w ValueWrapper) unwrapInto(path string, dstv reflect.Value, v sdktypes.Value) error {
	if w.Preunwrap != nil {
		v, err := w.Preunwrap(v)
		if err != nil {
			return err
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

	if handled, err := w.unwrapContainerInto(path, dstv, v); handled {
		return err
	}

	if handled, err := w.unwrapScalarInto(path, dstv, v); handled {
		return err
	}

	return fmt.Errorf("%scannot unwrap into %v", path, dstv.Type())
}

func (w ValueWrapper) unwrapPtrInto(path string, dstv reflect.Value, v sdktypes.Value) error {
	if dstv.CanSet() {
		if sdktypes.IsNothingValue(v) {
			// Target is a ptr and settable, and we need to
			// set nothing, so just set to nil.
			dstv.SetZero()
			return nil
		}

		if _, ok := dstv.Interface().(sdktypes.Value); ok {
			// Target is a Value, just assign it directly.
			dstv.Set(reflect.ValueOf(v))
			return nil
		}

		if _, ok := dstv.Interface().(sdktypes.Object); ok {
			// Target is an object which contains the actual value wrapped in a Value,
			// extract the actual value and assign it.

			vv := reflect.ValueOf(sdktypes.GetValue(v))

			if !vv.Type().AssignableTo(dstv.Type()) {
				return errors.New("cannot assign to non-value object")
			}

			dstv.Set(vv)
			return nil
		}
	}

	if dstv.IsNil() {
		if !dstv.CanSet() {
			return errors.New("dst is an unsettable nil ptr")
		}

		// Need to allocate a new ptr, to make elem settable.
		ptr := reflect.New(dstv.Type().Elem())
		dstv.Set(ptr)
	}

	return w.unwrapInto(path, dstv.Elem(), v)
}

func (w ValueWrapper) unwrapScalarInto(path string, dstv reflect.Value, v sdktypes.Value) (bool, error) {
	switch dstv.Interface().(type) {
	case time.Duration:
		d, err := sdktypes.ValueToDuration(v)
		if err != nil {
			return true, fmt.Errorf("%scannot convert duration", path)
		}

		dstv.Set(reflect.ValueOf(d))
		return true, nil

	case time.Time:
		t, err := sdktypes.ValueToTime(v)
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

		if sym, ok := u.(sdktypes.Symbol); ok {
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

func (w ValueWrapper) unwrapContainerInto(path string, dstv reflect.Value, v sdktypes.Value) (bool, error) {
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

	switch sdktypes.GetValue(v).(type) {
	case sdktypes.ListValue, sdktypes.SetValue:
		var isSet bool
		var vs []sdktypes.Value
		if sdktypes.IsListValue(v) {
			vs = sdktypes.GetListValue(v)
		} else {
			vs = sdktypes.GetSetValue(v)
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

	case sdktypes.DictValue, sdktypes.StructValue:
		var items []*sdktypes.DictValueItem

		if sdktypes.IsDictValue(v) {
			items = sdktypes.GetDictValue(v)
		} else {
			_, fields := sdktypes.GetStructValue(v)

			items = kittehs.TransformMapToList(fields, func(n string, v sdktypes.Value) *sdktypes.DictValueItem {
				return &sdktypes.DictValueItem{
					K: sdktypes.NewStringValue(n),
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
				if !sdktypes.IsStringValue(item.K) {
					return true, fmt.Errorf("%scannot convert non-string-key dict into struct", pathf(""))
				}

				k := sdktypes.GetStringValue(item.K)
				fn := w.toStructCaser(k)

				fv := sv.Elem().FieldByName(fn)
				if fv.Kind() == reflect.Invalid {
					if k != fn {
						fv = sv.Elem().FieldByName(k)
						if fv.Kind() == reflect.Invalid {
							return true, fmt.Errorf("%s field %q or %q does not exit", pathf(""), fn, k)
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
