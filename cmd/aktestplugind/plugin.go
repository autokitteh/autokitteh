package main

import (
	"context"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/pluginimpl"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/pluginsvc"
)

var Test = &pluginimpl.Plugin{
	ID:  "test",
	Doc: "test plugin",
	Members: map[string]*pluginimpl.PluginMember{
		"cat": pluginimpl.NewSimpleMethodMember(
			"returns cat's vocalization",
			func(context.Context, []*apivalues.Value, map[string]*apivalues.Value) (*apivalues.Value, error) {
				return apivalues.String("meow"), nil
			},
		),
		"dog": pluginimpl.NewSimpleMethodMember(
			"returns dog's vocalization",
			func(context.Context, []*apivalues.Value, map[string]*apivalues.Value) (*apivalues.Value, error) {
				return apivalues.String("woof"), nil
			},
		),
	},
}

func main() { pluginsvc.Run(Test) }
