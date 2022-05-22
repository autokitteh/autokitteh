package pluginimpl

import (
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
)

type FuncToValueFuncOpts struct{ Flags map[string]bool }

type FuncToValueFuncOptFunc func(*FuncToValueFuncOpts)

func WithFlags(flags ...string) FuncToValueFuncOptFunc {
	return func(opts *FuncToValueFuncOpts) {
		if opts.Flags == nil {
			opts.Flags = make(map[string]bool)
		}

		for _, flag := range flags {
			opts.Flags[flag] = true
		}
	}
}

// used to translate functions return values.
type FuncToValueFunc func(string, PluginMethodFunc, ...FuncToValueFuncOptFunc) *apivalues.Value
