package pluginimpl

import (
	"context"

	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
)

type PluginMethodFunc func(
	ctx context.Context,
	name string,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
	funcToValue FuncToValueFunc,
) (*apivalues.Value, error)

type SimplePluginMethodFunc func(
	ctx context.Context,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error)

type PluginMemberValue interface {
	isPluginMemberValue()
}

type PluginMember struct {
	Doc   string
	Value PluginMemberValue
}

type MethodPluginMember struct{ Func PluginMethodFunc }

func (m MethodPluginMember) isPluginMemberValue() {}

type ValuePluginMember struct{ *apivalues.Value }

func (m ValuePluginMember) isPluginMemberValue() {}

type LazyValuePluginMember struct {
	New func(FuncToValueFunc) *apivalues.Value
}

func (m LazyValuePluginMember) isPluginMemberValue() {}

func NewMethodMember(doc string, f PluginMethodFunc) *PluginMember {
	return &PluginMember{
		Doc:   doc,
		Value: MethodPluginMember{Func: f},
	}
}

func NewSimpleMethodMember(doc string, f SimplePluginMethodFunc) *PluginMember {
	return &PluginMember{
		Doc: doc,
		Value: MethodPluginMember{
			Func: func(
				ctx context.Context,
				_ string,
				args []*apivalues.Value,
				kwargs map[string]*apivalues.Value,
				_ FuncToValueFunc,
			) (*apivalues.Value, error) {
				return f(ctx, args, kwargs)
			},
		},
	}
}

func NewValueMember(doc string, v *apivalues.Value) *PluginMember {
	return &PluginMember{
		Doc:   doc,
		Value: ValuePluginMember{Value: v},
	}
}

func NewLazyValueMember(doc string, f func(FuncToValueFunc) *apivalues.Value) *PluginMember {
	return &PluginMember{
		Doc:   doc,
		Value: LazyValuePluginMember{New: f},
	}
}
