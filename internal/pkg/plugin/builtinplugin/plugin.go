package builtinplugin

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiplugin"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/plugin"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/pluginimpl"
)

var ErrNotACallValue = errors.New("not a call value")

type BuiltinPlugin struct {
	*pluginimpl.Plugin

	ID apiplugin.PluginID

	funcs funcs
}

func (p *BuiltinPlugin) Factory(id apiplugin.PluginID, pl *pluginimpl.Plugin) func() *BuiltinPlugin {
	return func() *BuiltinPlugin {
		return &BuiltinPlugin{Plugin: pl, ID: id}
	}
}

var _ plugin.Plugin = &BuiltinPlugin{}

func (p *BuiltinPlugin) Describe(ctx context.Context) (*apiplugin.PluginDesc, error) {
	return p.Desc(), nil
}

func (p *BuiltinPlugin) GetAll(ctx context.Context) (map[string]*apivalues.Value, error) {
	ret := make(map[string]*apivalues.Value, len(p.Members))

	for n := range p.Members {
		var err error
		if ret[n], err = p.Get(ctx, n); err != nil {
			return nil, fmt.Errorf("%q: %w", n, err)
		}
	}

	return ret, nil
}

func (p *BuiltinPlugin) Get(_ context.Context, name string) (*apivalues.Value, error) {
	m, ok := p.Members[name]
	if !ok {
		return nil, nil
	}

	// Ugly, but effective. this way we don't need special dance to initialize issuer before any call.
	p.funcs.issuer = p.ID.String()

	switch v := m.Value.(type) {
	case pluginimpl.LazyValuePluginMember:
		return v.New(p.funcs.AddUnique), nil

	case pluginimpl.ValuePluginMember:
		return v.Value, nil

	case pluginimpl.MethodPluginMember:
		return p.funcs.AddUnique(name, v.Func), nil

	default:
		panic("unknown member type")
	}
}

func (p *BuiltinPlugin) Call(ctx context.Context, v *apivalues.Value, args []*apivalues.Value, kwargs map[string]*apivalues.Value) (*apivalues.Value, error) {
	cv, ok := v.Get().(apivalues.CallValue)
	if !ok {
		return nil, ErrNotACallValue
	}

	f, err := p.funcs.Get(cv)
	if err != nil {
		return nil, err
	}

	return f(ctx, cv.Name, args, kwargs, p.funcs.AddDynamic)
}
