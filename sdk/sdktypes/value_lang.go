package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	valuev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
)

type StructValuePB = valuev1.Struct

type structValueTraits struct{ immutableObjectTrait }

func (structValueTraits) Validate(m *StructValuePB) error {
	return errors.Join(
		objectField[Value]("ctor", m.Ctor),
		valuesMapField("field", m.Fields),
	)
}

func (structValueTraits) StrictValidate(m *StructValuePB) error {
	return mandatory("ctor", m.Ctor)
}

var _ objectTraits[*StructValuePB] = structValueTraits{}

type StructValue struct {
	object[*StructValuePB, structValueTraits]
}

func init() { registerObject[StructValue]() }

func (StructValue) isConcreteValue() {}

func (s StructValue) Ctor() Value { return forceFromProto[Value](s.read().Ctor) }
func (s StructValue) Fields() map[string]Value {
	return kittehs.TransformMapValues(s.read().Fields, forceFromProto[Value])
}

func NewStructValue(ctor Value, fields map[string]Value) (Value, error) {
	return FromProto[Value](&ValuePB{Struct: &StructValuePB{
		Ctor:   ctor.ToProto(),
		Fields: kittehs.TransformMapValues(fields, ToProto),
	}})
}

func (v Value) IsStruct() bool         { return v.read().Struct != nil }
func (v Value) GetStruct() StructValue { return forceFromProto[StructValue](v.read().Struct) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsStruct() {
			return v.GetStruct()
		}
		return nil
	})
}

// ---

type ModuleValuePB = valuev1.Module

type moduleValueTraits struct{ immutableObjectTrait }

func (moduleValueTraits) Validate(m *ModuleValuePB) error {
	return nil
}

func (moduleValueTraits) StrictValidate(m *ModuleValuePB) error {
	return mandatory("name", m.Name)
}

var _ objectTraits[*ModuleValuePB] = moduleValueTraits{}

type ModuleValue struct {
	object[*ModuleValuePB, moduleValueTraits]
}

func init() { registerObject[ModuleValue]() }

func (ModuleValue) isConcreteValue() {}

func (s ModuleValue) Name() Symbol { return kittehs.Must1(ParseSymbol(s.read().Name)) }
func (s ModuleValue) Members() map[string]Value {
	return kittehs.TransformMapValues(s.read().Members, forceFromProto[Value])
}

func NewModuleValue(name Symbol, fields map[string]Value) (Value, error) {
	return FromProto[Value](&ValuePB{Module: &ModuleValuePB{
		Name:    name.String(),
		Members: kittehs.TransformMapValues(fields, ToProto),
	}})
}

func (v Value) IsModule() bool         { return v.read().Module != nil }
func (v Value) GetModule() ModuleValue { return forceFromProto[ModuleValue](v.read().Module) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsModule() {
			return v.GetModule()
		}
		return nil
	})
}
