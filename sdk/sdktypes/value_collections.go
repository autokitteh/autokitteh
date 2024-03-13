package sdktypes

import (
	"fmt"
	"slices"
	"sort"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	valuev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
)

type BytesValuePB = valuev1.Bytes

type BytesValue struct {
	object[*BytesValuePB, nopObjectTraits[*BytesValuePB]]
}

func (BytesValue) isConcreteValue() {}

func (s BytesValue) Value() []byte { return clone(s.m).V }

func NewBytesValue(v []byte) Value {
	return forceFromProto[Value](clone(&ValuePB{Bytes: &BytesValuePB{V: v}}))
}

func (v Value) IsBytes() bool        { return v.read().Bytes != nil }
func (v Value) GetBytes() BytesValue { return forceFromProto[BytesValue](v.read().Bytes) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsBytes() {
			return v.GetBytes()
		}
		return nil
	})
}

// ---

type ListValuePB = valuev1.List

type listValueTraits struct{}

func (listValueTraits) Validate(m *ListValuePB) error       { return valuesSlice(m.Vs) }
func (listValueTraits) StrictValidate(m *ListValuePB) error { return nil }

var _ objectTraits[*ListValuePB] = listValueTraits{}

type ListValue struct {
	object[*ListValuePB, listValueTraits]
}

func (ListValue) isConcreteValue() {}

func NewListValue(vs []Value) (Value, error) {
	return FromProto[Value](&ValuePB{List: &ListValuePB{Vs: kittehs.Transform(vs, ToProto)}})
}

func (l ListValue) Values() []Value { return kittehs.Transform(l.read().Vs, forceFromProto[Value]) }

func (v Value) IsList() bool       { return v.read().List != nil }
func (v Value) GetList() ListValue { return forceFromProto[ListValue](v.read().List) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsList() {
			return v.GetList()
		}
		return nil
	})
}

// ---

type SetValuePB = valuev1.Set

type setValueTraits struct{}

func (setValueTraits) Validate(m *SetValuePB) error       { return valuesSlice(m.Vs) }
func (setValueTraits) StrictValidate(m *SetValuePB) error { return nil }

var _ objectTraits[*SetValuePB] = setValueTraits{}

type SetValue struct {
	object[*SetValuePB, setValueTraits]
}

func (SetValue) isConcreteValue() {}

func NewSetValue(vs []Value) (Value, error) {
	vs = slices.Clone(vs)

	// Make item order deterministic.
	sort.Slice(vs, func(i, j int) bool { return vs[i].Hash() < vs[j].Hash() })

	return FromProto[Value](&ValuePB{Set: &SetValuePB{Vs: kittehs.Transform(vs, ToProto)}})
}

func (l SetValue) Values() []Value { return kittehs.Transform(l.read().Vs, forceFromProto[Value]) }

func (v Value) IsSet() bool      { return v.read().Set != nil }
func (v Value) GetSet() SetValue { return forceFromProto[SetValue](v.read().Set) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsSet() {
			return v.GetSet()
		}
		return nil
	})
}

// ---

type (
	DictValuePB = valuev1.Dict
	dictItemPB  = valuev1.Dict_Item
)

type dictValueTraits struct{}

func (dictValueTraits) Validate(m *DictValuePB) error {
	keys := make(map[string]*ValuePB, len(m.Items))

	for _, i := range m.Items {
		hash := hash(i.K)

		if keys[hash] != nil {
			return fmt.Errorf("duplicate key %q", i.K)
		}

		if err := validateValue(i.K); err != nil {
			return fmt.Errorf("key %q: %w", i.K, err)
		}

		keys[hash] = i.K

		if err := validateValue(i.V); err != nil {
			return fmt.Errorf("value for key %q: %w", i.K, err)
		}
	}

	return nil
}

func (dictValueTraits) StrictValidate(m *DictValuePB) error { return nil }

var _ objectTraits[*DictValuePB] = dictValueTraits{}

type DictValue struct {
	object[*DictValuePB, dictValueTraits]
}

type DictItem struct{ K, V Value }

func (DictValue) isConcreteValue() {}

func (d DictValue) Items() []DictItem {
	return kittehs.Transform(d.read().Items, func(i *dictItemPB) DictItem {
		return DictItem{K: forceFromProto[Value](i.K), V: forceFromProto[Value](i.V)}
	})
}

func (d DictValue) ToStringValuesMap() (map[string]Value, error) {
	return kittehs.ListToMapError(d.Items(), func(i DictItem) (string, Value, error) {
		s, err := i.K.ToString()
		if err != nil {
			return "", Value{}, err
		}
		return s, i.V, nil
	})
}

func NewDictValue(items []DictItem) (Value, error) {
	items = slices.Clone(items)

	// Make item order deterministic.
	sort.SliceStable(items, func(i, j int) bool { return items[i].K.Hash() < items[j].K.Hash() })

	return FromProto[Value](&ValuePB{Dict: &DictValuePB{
		Items: kittehs.Transform(items, func(i DictItem) *dictItemPB {
			return &dictItemPB{K: ToProto(i.K), V: ToProto(i.V)}
		}),
	}})
}

func NewDictValueFromStringMap(m map[string]Value) Value {
	return kittehs.Must1(NewDictValue(
		kittehs.TransformMapToList(
			m,
			func(k string, v Value) DictItem { return DictItem{K: NewStringValue(k), V: v} },
		),
	))
}

func (v Value) IsDict() bool       { return v.read().Dict != nil }
func (v Value) GetDict() DictValue { return forceFromProto[DictValue](v.read().Dict) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsDict() {
			return v.GetDict()
		}
		return nil
	})
}
