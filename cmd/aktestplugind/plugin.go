package main

import (
	"context"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/pluginimpl"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/pluginsvc"
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

func main() { pluginsvc.Run(nil, Test) }
