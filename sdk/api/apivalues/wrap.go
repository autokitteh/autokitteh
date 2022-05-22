package apivalues

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var varNameRegexp = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type wrapOpts struct {
	translate                func(interface{}) (interface{}, error)
	dictSorter               func([]*DictItem)
	listSorter               func([]*Value)
	structFieldNameConverter func(string) string
}

func WithStructFieldNameConverter(f func(string) string) func(*wrapOpts) {
	return func(o *wrapOpts) { o.structFieldNameConverter = f }
}

func WithDictSorter(s func([]*DictItem)) func(*wrapOpts) {
	return func(o *wrapOpts) { o.dictSorter = s }
}

func WithListSorter(s func([]*Value)) func(*wrapOpts) {
	return func(o *wrapOpts) { o.listSorter = s }
}

func WithWrapTranslate(t func(interface{}) (interface{}, error)) func(*wrapOpts) {
	return func(o *wrapOpts) { o.translate = t }
}

func WrapValuesMap(dst map[string]*Value, m map[string]interface{}, wopts ...func(*wrapOpts)) error {
	for k, v := range m {
		var err error
		if dst[k], err = Wrap(v, wopts...); err != nil {
			return fmt.Errorf("%q: %w", k, err)
		}
	}
	return nil
}

func WrapIntoValuesMap(dst map[string]*Value, v interface{}, wopts ...func(*wrapOpts)) error {
	wv, err := Wrap(v, wopts...)
	if err != nil {
		return err
	}

	switch vv := wv.Get().(type) {
	case DictValue:
		vv.ToStringValuesMap(dst)
		return nil
	case StructValue:
		for k, v := range vv.Fields {
			dst[k] = v
		}
		return nil
	case ModuleValue:
		for k, v := range vv.Members {
			dst[k] = v
		}
		return nil
	default:
		return fmt.Errorf("cannot wrap %v as map", reflect.TypeOf(wv.Get()))
	}
}

func MustWrap(v interface{}, wopts ...func(*wrapOpts)) *Value {
	w, err := Wrap(v, wopts...)
	if err != nil {
		panic(err)
	}
	return w
}

func Wrap(v interface{}, wopts ...func(*wrapOpts)) (*Value, error) {
	opts := wrapOpts{
		structFieldNameConverter: func(n string) string { return n },
	}

	for _, wopt := range wopts {
		wopt(&opts)
	}

	var err error

	if t := opts.translate; t != nil {
		if v, err = t(v); err != nil {
			return nil, err
		}
	}

	var val value

	if t, ok := v.(time.Time); ok {
		return NewValue(TimeValue(t))
	}

	if d, ok := v.(time.Duration); ok {
		return NewValue(DurationValue(d))
	}

	vv := reflect.ValueOf(v)
	switch vk := vv.Kind(); vk {
	case reflect.Invalid:
		val = NoneValue{}

	case reflect.Ptr:
		if !vv.IsNil() {
			return Wrap(vv.Elem().Interface(), wopts...)
		}

		val = NoneValue{}

	case reflect.Struct:
		vt := reflect.TypeOf(v)

		if vt.Size() != 0 {
			fs := make(map[string]*Value)

			for _, vfs := range reflect.VisibleFields(vt) {
				if !vfs.IsExported() || vfs.Anonymous {
					continue
				}

				jtag := vfs.Tag.Get("json")
				if jtag == "-" {
					continue
				}

				fv := vv.FieldByIndex(vfs.Index)

				wfv, err := Wrap(fv.Interface(), wopts...)
				if err != nil {
					return nil, fmt.Errorf("unable to convert struct field: %w", err)
				}

				n := vfs.Name

				if len(jtag) > 0 && jtag[0] != ',' {
					mn := strings.SplitN(jtag, ",", 2)[0]
					if varNameRegexp.MatchString(mn) {
						n = mn
					}
				}

				n = opts.structFieldNameConverter(n)

				fs[n] = wfv
			}

			n := vt.Name()
			if len(n) == 0 {
				n = "struct"
			}

			return NewValue(
				StructValue{
					Ctor:   Symbol(n),
					Fields: fs,
				},
			)
		}

		val = NoneValue{}

	case reflect.Array, reflect.Slice:
		if vv.Type().Elem().Kind() == reflect.Uint8 {
			val = BytesValue(vv.Interface().([]byte))
			break
		}

		vs := make([]*Value, vv.Len())
		for i := 0; i < vv.Len(); i++ {
			if vs[i], err = Wrap(vv.Index(i).Interface(), wopts...); err != nil {
				return nil, fmt.Errorf("%d: %w", i, err)
			}
		}

		if s := opts.listSorter; s != nil {
			s(vs)
		}

		val = ListValue(vs)

	case reflect.Map:
		vs := make([]*DictItem, 0, vv.Len())
		for i := vv.MapRange(); i.Next(); {
			k, v := i.Key(), i.Value()

			di := &DictItem{}

			if di.K, err = Wrap(k.Interface(), wopts...); err != nil {
				return nil, fmt.Errorf("key %v: %w", k, err)
			}

			if di.V, err = Wrap(v.Interface(), wopts...); err != nil {
				return nil, fmt.Errorf("key %v: %w", k, err)
			}

			vs = append(vs, di)
		}

		if s := opts.dictSorter; s != nil {
			s(vs)
		}

		val = DictValue(vs)

	default:
		switch vv := v.(type) {
		case bool:
			val = BooleanValue(vv)
		case int:
			val = IntegerValue(vv)
		case int8:
			val = IntegerValue(vv)
		case int16:
			val = IntegerValue(vv)
		case int32:
			val = IntegerValue(vv)
		case int64:
			val = IntegerValue(vv)
		case string:
			val = StringValue(vv)
		case float32:
			val = FloatValue(vv)
		case float64:
			val = FloatValue(vv)
		case json.Number:
			if i64, err := vv.Int64(); err == nil {
				val = IntegerValue(i64)
			} else if f64, err := vv.Float64(); err == nil {
				val = FloatValue(f64)
			} else {
				val = StringValue(vv)
			}
		default:
			return nil, fmt.Errorf("unhandled type: %q/%q", vk, reflect.TypeOf(v))
		}
	}

	return NewValue(val)
}
