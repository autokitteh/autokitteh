package sdktypes

import "errors"

type MapKind int

const (
	MapKindValue                 = iota
	MapKindDictItemKey   MapKind = iota
	MapKindDictItemValue MapKind = iota
	MapKindStructField
)

type MapInfo struct {
	Kind       MapKind
	FieldName  string
	Key, Value Value
}

var ErrMapSkip = errors.New("skip")

// f must return a valid value if no error.
func (v Value) Map(f func(v Value, info *MapInfo) (Value, error)) (Value, error) {
	return v.map_(nil, f)
}

func (v Value) map_(info *MapInfo, f func(v Value, info *MapInfo) (Value, error)) (Value, error) {
	if info == nil {
		info = &MapInfo{}
	}

	v, err := f(v, info)
	if err == ErrMapSkip {
		return v, nil
	}

	if err != nil {
		return InvalidValue, err
	}

	switch vv := v.Concrete().(type) {
	case ListValue:
		var vs []Value
		for _, v := range vv.Values() {
			v, err := v.map_(nil, f)
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
			v, err := v.map_(nil, f)
			if err != nil {
				return InvalidValue, err
			} else if !v.IsValid() {
				return InvalidValue, errors.New("invalid value")
			}
		}
		return NewSetValue(vs)

	case DictValue:
		items := vv.Items()
		for i := range items {
			if items[i].K, err = items[i].K.map_(&MapInfo{Key: items[i].K, Value: items[i].V, Kind: MapKindDictItemKey}, f); err != nil {
				return InvalidValue, err
			} else if !items[i].K.IsValid() {
				return InvalidValue, errors.New("invalid value")
			}

			if items[i].V, err = items[i].V.map_(&MapInfo{Key: items[i].K, Value: items[i].V, Kind: MapKindDictItemValue}, f); err != nil {
				return InvalidValue, err
			} else if !items[i].V.IsValid() {
				return InvalidValue, errors.New("invalid value")
			}
		}
		return NewDictValue(items)

	case StructValue:
		fs := vv.Fields()
		for k, fv := range fs {
			if fs[k], err = fv.map_(&MapInfo{FieldName: k, Kind: MapKindStructField}, f); err != nil {
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
