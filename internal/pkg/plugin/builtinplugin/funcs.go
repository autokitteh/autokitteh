package builtinplugin

import (
	"fmt"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/pluginimpl"
)

type funcs struct {
	issuer      string
	i           uint64
	idToFunc    map[string]pluginimpl.PluginMethodFunc
	nameToValue map[string]*apivalues.Value // for unique values only.
}

// Must be called in lock context.
func (c *funcs) add(name string, f pluginimpl.PluginMethodFunc, fopts ...pluginimpl.FuncToValueFuncOptFunc) *apivalues.Value {
	var opts pluginimpl.FuncToValueFuncOpts
	for _, fopt := range fopts {
		fopt(&opts)
	}

	c.i++

	id := fmt.Sprintf("C%4.4x", c.i)

	if c.idToFunc == nil {
		c.idToFunc = make(map[string]pluginimpl.PluginMethodFunc, 16)
	}

	if c.idToFunc[id] != nil {
		panic(id)
	}

	c.idToFunc[id] = f

	return apivalues.MustNewValue(apivalues.CallValue{ID: id, Name: name, Flags: opts.Flags, Issuer: c.issuer})
}

// This must be called only for unique funcs (plugin members) and not for return values.
// This function makes sure that the same value will be returned every time, as the function
// supplied is not generated within another function's scope.
func (c *funcs) AddUnique(name string, f pluginimpl.PluginMethodFunc, opts ...pluginimpl.FuncToValueFuncOptFunc) *apivalues.Value {
	if v, ok := c.nameToValue[name]; ok {
		return v
	}

	v := c.AddDynamic(name, f, opts...)

	if c.nameToValue == nil {
		c.nameToValue = make(map[string]*apivalues.Value, 16)
	}

	c.nameToValue[name] = v

	return v
}

// Use this for dynamically generated values.
func (c *funcs) AddDynamic(name string, f pluginimpl.PluginMethodFunc, opts ...pluginimpl.FuncToValueFuncOptFunc) *apivalues.Value {
	return c.add(name, f, opts...)
}

var (
	nilfuncs *funcs
	_, _     pluginimpl.FuncToValueFunc = nilfuncs.AddDynamic, nilfuncs.AddUnique
)

func (c *funcs) Get(cv apivalues.CallValue) (pluginimpl.PluginMethodFunc, error) {
	f, ok := c.idToFunc[cv.ID]
	if !ok {
		return nil, fmt.Errorf("value not registered as a function")
	}

	return f, nil
}
