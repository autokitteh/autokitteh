package sdktypes

import (
	"reflect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
)

// Helper to easily access list of Vars.

type Vars []Var

func NewVars(vs ...Var) Vars { return vs }

func (vs Vars) Append(v Var) Vars { return append(vs, v) }

// panics if n is an invalid var name.
func (vs Vars) Set(n string, v string, isSecret bool) Vars {
	return vs.Append(NewVar(forceSymbol(n), v, isSecret))
}

func (vs Vars) GetValue(name Symbol) string { return vs.Get(name).String() }

func (vs Vars) GetValueByString(name string) string { return vs.GetByString(name).String() }

func (vs Vars) Get(name Symbol) Var { return vs.GetByString(name.String()) }

func (vs Vars) GetByString(name string) Var {
	_, v := kittehs.FindFirst(vs, func(v Var) bool { return v.Name().String() == name })
	return v
}

func (vs Vars) Has(name Symbol) bool { return vs.Get(name).IsValid() }

func (vs Vars) ToStringMap() map[string]string {
	return kittehs.ListToMap(vs, func(v Var) (string, string) {
		return v.Name().String(), v.Value()
	})
}

func (vs Vars) ToMap() map[Symbol]Var {
	return kittehs.ListToMap(vs, func(v Var) (Symbol, Var) { return v.Name(), v })
}

// Encodes `in` into Vars. `in` must be a struct or a non-nil pointer to a struct.
// All members must be strings. A field tag of `secret` will make the field secret.
func EncodeVars(in any) (vs Vars) {
	v, t := reflect.ValueOf(in), reflect.TypeOf(in)

	if t.Kind() == reflect.Ptr {
		v, t = v.Elem(), t.Elem()
	}

	if v.Kind() != reflect.Struct {
		sdklogger.Panic("invalid type - must be a struct")
	}

	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		n := forceSymbol(ft.Name)

		if ft.Type.Kind() != reflect.String {
			sdklogger.Panic("invalid field value type - not a string")
		}

		tag := ft.Tag.Get("var")

		v := fv.Interface().(string)

		vs = vs.Append(NewVar(n, v, tag == "secret"))
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

	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		n := forceSymbol(ft.Name)

		if ft.Type.Kind() != reflect.String {
			sdklogger.Panic("invalid field value type - not a string")
		}

		v := vs.Get(n).Value()

		fv.SetString(v)
	}
}
