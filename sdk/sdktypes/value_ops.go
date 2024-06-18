package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

func (v Value) IsTruthy() bool {
	if !v.IsValid() {
		return false
	}

	return v.Concrete().IsTrue()
}

func (v Value) Items() ([]Value, error) {
	switch v := v.Concrete().(type) {
	case ListValue:
		return v.Values(), nil
	case SetValue:
		return v.Values(), nil
	case DictValue:
		return kittehs.Transform(v.Items(), func(kv DictItem) Value { return kv.K }), nil
	default:
		return nil, errors.New("value is not a collection")
	}
}

func (v Value) Len() (int, error) {
	switch v := v.Concrete().(type) {
	case ListValue:
		return v.Len(), nil
	case SetValue:
		return v.Len(), nil
	case DictValue:
		return v.Len(), nil
	default:
		return 0, errors.New("value is not a collection")
	}
}

func (v Value) Index(i int) (Value, error) {
	switch v := v.Concrete().(type) {
	case ListValue:
		return v.Index(i)
	case SetValue:
		return v.Index(i)
	case DictValue:
		return v.Index(i)
	default:
		return InvalidValue, errors.New("value is not a collection")
	}
}

func (v Value) SetKey(k Value, vv Value) (Value, error) {
	switch v := v.Concrete().(type) {
	case DictValue:
		items := v.Items()
		var found bool
		for i, item := range items {
			if item.K.Equal(k) {
				items[i].V = vv
				found = true
				break
			}
		}
		if !found {
			items = append(items, DictItem{K: k, V: vv})
		}
		return NewDictValue(items)
	case StructValue:
		if vv.IsString() {
			sym, err := StrictParseSymbol(vv.GetString().Value())
			if err != nil {
				return InvalidValue, err
			}

			vv = NewSymbolValue(sym)
		}

		if !vv.IsSymbol() {
			return InvalidValue, errors.New("value is not a symbol")
		}
		fs := v.Fields()
		if fs == nil {
			fs = make(map[string]Value, 1)
		}
		fs[k.String()] = vv
		return NewStructValue(v.Ctor(), fs)
	default:
		return InvalidValue, errors.New("value is not a collection")
	}
}

func (v Value) GetKey(k Value) (Value, error) {
	switch v := v.Concrete().(type) {
	case DictValue:
		items := v.Items()
		_, item := kittehs.FindFirst(items, func(item DictItem) bool { return item.K.Equal(k) })
		if item.K.IsValid() {
			return item.V, nil
		}
		return InvalidValue, nil

	case StructValue:
		if k.IsString() {
			sym, err := StrictParseSymbol(k.GetString().Value())
			if err != nil {
				return InvalidValue, err
			}

			k = NewSymbolValue(sym)
		}

		if !k.IsSymbol() {
			return InvalidValue, errors.New("key is not a symbol")
		}

		return v.Fields()[kittehs.Must1(k.ToString())], nil
	default:
		return InvalidValue, errors.New("value is not a collection")
	}
}

func (v Value) Append(vv Value) (Value, error) {
	if !v.IsValid() {
		return NewListValue([]Value{vv})
	}

	switch v := v.Concrete().(type) {
	case ListValue:
		return NewListValue(append(v.Values(), vv))
	case SetValue:
		return NewSetValue(append(v.Values(), vv))
	default:
		return InvalidValue, errors.New("value is not a list")
	}
}
