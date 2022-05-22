package apivalues

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

type unwrapper interface {
	UnwrapInto(interface{}) (bool, error)
}

type unwrapOpts struct {
	jsonSafe bool
}

func WithUnwrapJSONSafe() func(*unwrapOpts) { return func(o *unwrapOpts) { o.jsonSafe = true } }

func UnwrapValuesMap(m map[string]*Value, fopts ...func(*unwrapOpts)) map[string]interface{} {
	ret := make(map[string]interface{}, len(m))
	for k, v := range m {
		ret[k] = v.Unwrap(fopts...)
	}
	return ret
}

func UnwrapInto(dst interface{}, v value, fopts ...func(*unwrapOpts)) error {
	if reflect.TypeOf(dst).Kind() != reflect.Ptr {
		panic("dst is not a ptr")
	}

	if vv, ok := dst.(**Value); ok {
		*vv = MustNewValue(v)
		return nil
	}

	if cvt, ok := dst.(converter); ok {
		// Maybe destination supply convertion.
		if err := cvt.ConvertFrom(v); err == nil {
			return nil
		} else if !errors.Is(err, ErrNoConversion) {
			return err
		}

		// fallthrough if ErrNoConversion.
	}

	if unw, ok := v.(unwrapper); ok {
		if fit, err := unw.UnwrapInto(dst); err != nil {
			return err
		} else if fit {
			return nil
		}
	}

	dstv := reflect.ValueOf(dst).Elem()

	if !dstv.CanSet() {
		panic("dst is not settable")
	}

	srcv := reflect.ValueOf(Unwrap(v, fopts...))

	if !srcv.Type().AssignableTo(dstv.Type()) {
		if !srcv.CanConvert(dstv.Type()) {
			return fmt.Errorf("%q is not assignable to %q", srcv.Type(), dstv.Type())
		}

		srcv = srcv.Convert(dstv.Type())
	}

	// TODO: this will only work for the simplest cases.
	//       will not work for map[interface{}]interface{} and such.
	dstv.Set(srcv)

	return nil
}

func Unwrap(v value, fopts ...func(*unwrapOpts)) interface{} {
	var opts unwrapOpts
	for _, fopt := range fopts {
		fopt(&opts)
	}

	switch vv := v.(type) {
	case NoneValue:
		return struct{}{}
	case StringValue:
		return string(vv)
	case IntegerValue:
		return int64(vv)
	case FloatValue:
		return float32(vv)
	case BooleanValue:
		return bool(vv)
	case BytesValue:
		return []byte(vv)
	case TimeValue:
		return time.Time(vv)
	case DurationValue:
		return time.Duration(vv)
	case ListValue:
		vs := make([]interface{}, len(vv))
		for i, v := range vv {
			vs[i] = Unwrap(v.Get(), fopts...)
		}
		return vs
	case SetValue:
		vs := make([]interface{}, len(vv))
		for i, v := range vv {
			vs[i] = v.Unwrap(fopts...)
		}
		return vs
	case DictValue:
		if opts.jsonSafe {
			vs := make(map[string]interface{}, len(vv))
			for _, kv := range vv {
				vs[kv.K.String()] = kv.V.Unwrap(fopts...)
			}
			return vs
		}

		vs := make(map[interface{}]interface{}, len(vv))
		for _, kv := range vv {
			vs[kv.K.Unwrap(fopts...)] = kv.V.Unwrap(fopts...)
		}
		return vs
	case CallValue, StructValue, ModuleValue, FunctionValue, SymbolValue:
		return v
	default:
		panic(fmt.Errorf("unrecognized type: %v", reflect.TypeOf(v)))
	}
}
