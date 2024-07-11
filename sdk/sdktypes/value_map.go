package sdktypes

import "errors"

// f must return a valid value if no error.
func (v Value) Map(f func(v Value) (Value, error)) (Value, error) {
	v, err := f(v)
	if err != nil {
		return InvalidValue, err
	}

	switch vv := v.Concrete().(type) {
	case ListValue:
		var vs []Value
		for _, v := range vv.Values() {
			v, err := v.Map(f)
			if err != nil {
				return InvalidValue, err
			} else if v.IsValid() {
				vs = append(vs, v)
			}
		}
		return NewListValue(vs)

	case SetValue:
		var vs []Value
		for _, v := range vv.Values() {
			v, err := v.Map(f)
			if err != nil {
				return InvalidValue, err
			} else if !v.IsValid() {
				return InvalidValue, errors.New("invalid value")
			}
		}
		return NewSetValue(vs)

	case DictValue:
		items := vv.Items()

		for _, item := range items {
			if item.K, err = item.K.Map(f); err != nil {
				return InvalidValue, err
			} else if !item.K.IsValid() {
				return InvalidValue, errors.New("invalid value")
			}

			if item.V, err = item.V.Map(f); err != nil {
				return InvalidValue, err
			} else if !item.V.IsValid() {
				return InvalidValue, errors.New("invalid value")
			}
		}

		return NewDictValue(items)

	case StructValue:
		fs := vv.Fields()

		for k, fv := range fs {
			if fs[k], err = fv.Map(f); err != nil {
				return InvalidValue, err
			} else if !fs[k].IsValid() {
				return InvalidValue, errors.New("invalid value")
			}
		}

		return NewStructValue(vv.Ctor(), fs)

	default:
		return v, nil
	}
}
