package sdktypes

import (
	"reflect"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
)

// Helper to easily access list of Vars.

type Vars []Var

func NewVars(vs ...Var) Vars { return vs }

func (vs Vars) WithPrefix(prefix string) Vars {
	varsWithPrefix := make(Vars, len(vs))
	for i, v := range vs {
		varsWithPrefix[i] = NewVar(NewSymbol(prefix + v.Name().String())).
			SetValue(v.Value()).SetSecret(v.IsSecret())
	}
	return varsWithPrefix
}

func (vs Vars) WithScopeID(vsid VarScopeID) Vars {
	varsWithScopeID := make(Vars, len(vs))
	for i, v := range vs {
		varsWithScopeID[i] = v.WithScopeID(vsid)
	}
	return varsWithScopeID
}

func (vs Vars) Append(others ...Var) Vars { return append(vs, others...) }

// panics if n is an invalid var name.
func (vs Vars) Set(n Symbol, v string, isSecret bool) Vars {
	return vs.Append(NewVar(n).SetSecret(isSecret).SetValue(v))
}

func (vs Vars) GetValue(name Symbol) string { return vs.Get(name).Value() }

func (vs Vars) GetValueByString(name string) string { return vs.GetByString(name).Value() }

func (vs Vars) Get(name Symbol) Var { return vs.GetByString(name.String()) }

func (vs Vars) GetByString(name string) Var {
	_, v := kittehs.FindFirst(vs, func(v Var) bool { return v.Name().String() == name })
	return v
}

func (vs Vars) Has(name Symbol) bool { return vs.Get(name).IsValid() }

func (vs Vars) ToMap() map[Symbol]Var {
	return kittehs.ListToMap(vs, func(v Var) (Symbol, Var) { return v.Name(), v })
}

// EncodeVars encodes a struct of strings (or a non-nil pointer to it) into Vars.
// Any other input would cause a panic. Optional field tags can be used to rename
// the fields and/or to mark them as secret: `var:"[new_name,]secret"`.
func EncodeVars(in any) (vs Vars) {
	v, t := reflect.ValueOf(in), reflect.TypeOf(in)

	if t.Kind() == reflect.Ptr {
		v, t = v.Elem(), t.Elem()
	}

	if v.Kind() != reflect.Struct {
		sdklogger.Panic("invalid type - must be a struct")
	}

	for i := range v.NumField() {
		fv := v.Field(i)
		ft := t.Field(i)

		if ft.Type.Kind() != reflect.String {
			sdklogger.Panic("invalid field value type - not a string")
		}

		// Guaranteed to have at least one element, even if it's empty
		// ("" -> [""], "x" -> ["x"], "x,y" -> ["x", "y"]).
		tag := strings.Split(ft.Tag.Get("var"), ",")

		n := NewSymbol(ft.Name)
		if tag[0] != "" && (tag[0] != "secret" || len(tag) > 1) {
			n = NewSymbol(tag[0])
		}

		v := fv.Interface().(string)
		isSecret := tag[len(tag)-1] == "secret"

		vs = vs.Append(NewVar(n).SetValue(v).SetSecret(isSecret))
	}

	return
}

// Decode Vars into `out`. `out` must be a non-nil pointer to a struct.
func (vs Vars) Decode(out any) {
	v, t := reflect.ValueOf(out), reflect.TypeOf(out)

	if t.Kind() != reflect.Ptr {
		sdklogger.Panic("invalid type - must be a pointer")
	}

	v, t = v.Elem(), t.Elem()

	if t.Kind() != reflect.Struct {
		sdklogger.Panic("invalid type - must be a struct")
	}

	for i := range v.NumField() {
		fv := v.Field(i)
		ft := t.Field(i)

		n := NewSymbol(ft.Name)

		if ft.Type.Kind() != reflect.String {
			sdklogger.Panic("invalid field value type - not a string")
		}

		v := vs.Get(n).Value()

		fv.SetString(v)
	}
}
