package internalplugins

import (
	"context"
	"os"

	"go.autokitteh.dev/sdk/api/apivalues"
	"go.autokitteh.dev/sdk/pluginimpl"
)

var OS = &pluginimpl.Plugin{
	Doc: "TODO",
	Members: map[string]*pluginimpl.PluginMember{
		"getenv": pluginimpl.NewSimpleMethodMember(
			"TODO",
			func(
				ctx context.Context,
				args []*apivalues.Value,
				kwargs map[string]*apivalues.Value,
			) (*apivalues.Value, error) {
				var name, def string

				if err := pluginimpl.UnpackArgs(
					args, kwargs,
					"name", &name,
					"def?", &def,
				); err != nil {
					return nil, err
				}

				v, found := os.LookupEnv(name)
				if !found {
					v = def
				}

				return apivalues.String(v), nil
			},
		),
		"exec": pluginimpl.NewSimpleMethodMember(
			"TODO",
			func(
				ctx context.Context,
				args []*apivalues.Value,
				kwargs map[string]*apivalues.Value,
			) (*apivalues.Value, error) {
				var (
					fail         = true
					path, dir    string
					cmdargs, env []interface{}
				)

				if err := pluginimpl.UnpackArgs(
					args, kwargs,
					"path?", &path,
					"args?", &cmdargs,
					"env?", &env,
					"dir?", &dir,
					"error?", &fail,
				); err != nil {
					return nil, err
				}

				return run(ctx, path, cmdargs, env, dir, fail)
			},
		),
		"shell": pluginimpl.NewSimpleMethodMember(
			"TODO",
			func(
				ctx context.Context,
				args []*apivalues.Value,
				kwargs map[string]*apivalues.Value,
			) (*apivalues.Value, error) {
				var (
					fail     = true
					cmd, dir string
					shell    = "/bin/sh"
					env      []interface{}
				)

				if err := pluginimpl.UnpackArgs(
					args, kwargs,
					"cmd", &cmd,
					"shell?", &shell,
					"env?", &env,
					"dir?", &dir,
					"error?", &fail,
				); err != nil {
					return nil, err
				}

				return run(ctx, shell, []interface{}{"-c", cmd}, env, dir, fail)
			},
		),
	},
}
