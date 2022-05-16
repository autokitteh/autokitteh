package pluginimpl

import (
	"context"
	"fmt"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
)

func SimplifyPluginMethodFunc(f func(context.Context, []*apivalues.Value, map[string]*apivalues.Value) (*apivalues.Value, error)) PluginMethodFunc {
	return func(
		ctx context.Context,
		_ string,
		args []*apivalues.Value,
		kwargs map[string]*apivalues.Value,
		_ FuncToValueFunc,
	) (*apivalues.Value, error) {
		return f(ctx, args, kwargs)
	}
}

type StructMember struct {
	*PluginMember
	Name string
}

func NewStructFuncMember(name, doc string, f PluginMethodFunc) *StructMember {
	return &StructMember{
		Name:         name,
		PluginMember: NewMethodMember(doc, f),
	}
}

func NewStructSimpleFuncMember(name, doc string, f SimplePluginMethodFunc) *StructMember {
	return &StructMember{
		Name:         name,
		PluginMember: NewSimpleMethodMember(doc, f),
	}
}

func NewStructValueMember(name, doc string, v *apivalues.Value) *StructMember {
	return &StructMember{
		Name:         name,
		PluginMember: NewValueMember(doc, v),
	}
}

func MustBuildStruct(funcToValue FuncToValueFunc, name string, members ...*StructMember) *apivalues.Value {
	v, err := BuildStruct(funcToValue, name, members...)
	if err != nil {
		panic(err)
	}
	return v
}

func BuildStruct(funcToValue FuncToValueFunc, name string, members ...*StructMember) (*apivalues.Value, error) {
	vmembers := make(map[string]*apivalues.Value, len(members))

	for _, m := range members {
		switch v := m.Value.(type) {
		case MethodPluginMember:
			vmembers[m.Name] = funcToValue(fmt.Sprintf("%s.%s", name, m.Name), v.Func)
		case ValuePluginMember:
			vmembers[m.Name] = v.Value
		default:
			panic("unknown member type")
		}
	}

	return apivalues.Struct(apivalues.Symbol(name), vmembers), nil
}
