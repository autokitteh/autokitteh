package internalplugins

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/sdk/pluginimpl"
)

func run(ctx context.Context, path string, args, env []interface{}, dir string, fail bool) (*apivalues.Value, error) {
	sargs := make([]string, len(args))
	for i, a := range args {
		s, ok := a.(string)
		if !ok {
			return nil, fmt.Errorf("args must be a list of strings")
		}

		sargs[i] = s
	}

	cmd := exec.CommandContext(ctx, path, sargs...)
	cmd.Dir = dir

	if env != nil {
		cmd.Env = make([]string, len(env))
		for i, e := range env {
			s, ok := e.(string)
			if !ok {
				return nil, fmt.Errorf("env must be a list of strings")
			}

			cmd.Env[i] = s
		}
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ee := err.(*exec.ExitError); fail || !ee {
			return nil, err
		}
	}

	return apivalues.MustNewValue(apivalues.ListValue(
		[]*apivalues.Value{
			apivalues.Integer(int64(cmd.ProcessState.ExitCode())),
			apivalues.String(string(out)),
		},
	)), nil
}

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
