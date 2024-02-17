package sdkmodule

import (
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type funcOpts struct {
	desc  sdktypes.ModuleFunctionPB
	flags []sdktypes.FunctionFlag
	fn    sdkexecutor.Function
}

type FuncOpt func(*funcOpts) error

func WithFuncDesc(desc string) FuncOpt {
	return func(cfg *funcOpts) error {
		cfg.desc.Description = desc
		return nil
	}
}

func WithFuncDoc(doc string) FuncOpt {
	return func(cfg *funcOpts) error {
		cfg.desc.DocumentationUrl = doc
		return nil
	}
}

func WithArg(arg string) FuncOpt { return WithArgs(arg) }

func WithArgs(args ...string) FuncOpt {
	return func(cfg *funcOpts) error {
		cfg.desc.Input = append(cfg.desc.Input, kittehs.Transform(args, func(s string) *sdktypes.ModuleFunctionFieldPB {
			return &sdktypes.ModuleFunctionFieldPB{
				Name:     strings.TrimRight(s, "?="),
				Kwarg:    strings.Contains(s, "="),
				Optional: strings.Contains(s, "?"),
			}
		})...)
		return nil
	}
}

func WithFlag(flag sdktypes.FunctionFlag) FuncOpt { return WithFlags(flag) }

func WithFlags(flags ...sdktypes.FunctionFlag) FuncOpt {
	return func(cfg *funcOpts) error {
		cfg.flags = append(cfg.flags, flags...)
		return nil
	}
}
