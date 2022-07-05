package internalplugins

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

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
		"read_text_file": pluginimpl.NewSimpleMethodMember(
			"TODO",
			func(
				ctx context.Context,
				args []*apivalues.Value,
				kwargs map[string]*apivalues.Value,
			) (*apivalues.Value, error) {
				var name string

				if err := pluginimpl.UnpackArgs(
					args, kwargs,
					"name", &name,
				); err != nil {
					return nil, err
				}

				bs, err := os.ReadFile(name)
				if err != nil {
					return nil, err
				}

				return apivalues.String(string(bs)), nil
			},
		),
		"write_text_file": pluginimpl.NewSimpleMethodMember(
			"TODO",
			func(
				ctx context.Context,
				args []*apivalues.Value,
				kwargs map[string]*apivalues.Value,
			) (*apivalues.Value, error) {
				var (
					name, text, pattern string
					mode                int64 = 0644
				)

				if err := pluginimpl.UnpackArgs(
					args, kwargs,
					"name?", &name,
					"text", &text,
					"mode?", &mode,
					"pattern?", &pattern,
				); err != nil {
					return nil, err
				}

				if name == "" {
					f, err := os.CreateTemp("", pattern)
					if err != nil {
						return nil, fmt.Errorf("create temp: %w", err)
					}

					defer f.Close()

					name = f.Name()

					if _, err := f.Write([]byte(text)); err != nil {
						return nil, fmt.Errorf("write: %w", err)
					}
				} else {
					if err := os.WriteFile(name, []byte(text), os.FileMode(mode)); err != nil {
						return nil, err
					}
				}

				return apivalues.String(name), nil
			},
		),
		"make_temp_dir": pluginimpl.NewSimpleMethodMember(
			"TODO",
			func(
				ctx context.Context,
				args []*apivalues.Value,
				kwargs map[string]*apivalues.Value,
			) (*apivalues.Value, error) {
				var pattern string

				if err := pluginimpl.UnpackArgs(
					args, kwargs,
					"pattern?", &pattern,
				); err != nil {
					return nil, err
				}

				path, err := os.MkdirTemp("", pattern)
				if err != nil {
					return nil, err
				}

				return apivalues.String(path), nil
			},
		),
		"look_path": pluginimpl.NewSimpleMethodMember(
			"TODO",
			func(
				ctx context.Context,
				args []*apivalues.Value,
				kwargs map[string]*apivalues.Value,
			) (*apivalues.Value, error) {
				var file string

				if err := pluginimpl.UnpackArgs(
					args, kwargs,
					"file", &file,
				); err != nil {
					return nil, err
				}

				path, err := exec.LookPath(file)
				if err != nil {
					if errors.Is(err, exec.ErrNotFound) {
						return apivalues.None, nil
					}

					return nil, err
				}

				return apivalues.String(path), nil
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
					"path", &path,
					"args?", &cmdargs,
					"env?", &env,
					"dir?", &dir,
					"fail?", &fail,
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
					"fail?", &fail,
				); err != nil {
					return nil, err
				}

				return run(ctx, shell, []interface{}{"-c", cmd}, env, dir, fail)
			},
		),
	},
}
