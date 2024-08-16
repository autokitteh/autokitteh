package runtimes

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	configruntimesvc "go.autokitteh.dev/autokitteh/runtimes/configrt/runtimesvc"
	pythonruntime "go.autokitteh.dev/autokitteh/runtimes/pythonrt"
	starlarkruntimesvc "go.autokitteh.dev/autokitteh/runtimes/starlarkrt/runtimesvc"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkruntimesclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	// Command separated list of runtimes and their addresses. If no address is supplied,
	// the runtime is builtin.
	// Examples:
	// - select runtimes: python=localhost:9980,starlark=localhost:9981,config
	// - all builtins: *
	// - all runtimes at remote: *=localhost:9981
	Runtimes string `koanf:"runtimes"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
	Dev:     &Config{},
}

var (
	builtins = []*sdkruntimes.Runtime{
		starlarkruntimesvc.Runtime,
		configruntimesvc.Runtime,
		pythonruntime.Runtime,
	}

	builtinsMap = kittehs.ListToMap(
		builtins,
		func(rt *sdkruntimes.Runtime) (string, *sdkruntimes.Runtime) { return rt.Desc.Name().String(), rt },
	)
)

func New(sl *zap.SugaredLogger, cfg *Config) (sdkservices.Runtimes, error) {
	ctx := context.Background()

	if sl == nil {
		sl = zap.NewNop().Sugar()
	}

	var rts []*sdkruntimes.Runtime
	names := kittehs.ListToMap(rts, func(rt *sdkruntimes.Runtime) (string, bool) { return rt.Desc.Name().String(), true })

	if cfg == nil {
		cfg = &Config{}
	}

	if cfg.Runtimes == "" {
		cfg.Runtimes = "*"
	}

	remotes := kittehs.Transform(strings.Split(cfg.Runtimes, ","), strings.TrimSpace)
	for _, remote := range remotes {
		add := func(rtsvc sdkservices.Runtimes, name, addr string) error {
			if names[name] {
				return fmt.Errorf("duplicate runtime found: %q", name)
			}

			names[name] = true

			if addr == "" {
				if rt := builtinsMap[name]; rt != nil {
					sl.Infof("using builtin runtime: %q", name)
					rts = append(rts, rt)
					return nil
				}

				return fmt.Errorf("invalid builtin runtime")
			}

			sym, err := sdktypes.ParseSymbol(name)
			if err != nil {
				return fmt.Errorf("invalid runtime name: %w", err)
			}

			rt, err := rtsvc.New(ctx, sym)
			if err != nil {
				return fmt.Errorf("failed to create runtime %q: %w", name, err)
			}

			sl.Infof("using remote runtime: %q at %s", name, addr)

			rts = append(rts, &sdkruntimes.Runtime{
				Desc: rt.Get(),
				New:  func() (sdkservices.Runtime, error) { return rt, nil },
			})

			return nil
		}

		name, addr, _ := strings.Cut(remote, "=")

		var rtsvc sdkservices.Runtimes

		if addr == "" {
			if name == "*" {
				// all builtins.
				for k := range builtinsMap {
					if err := add(nil, k, ""); err != nil {
						return nil, fmt.Errorf("runtime %q: %w", name, err)
					}
				}

				continue
			}

			// specific builtin.
			if err := add(nil, name, ""); err != nil {
				return nil, fmt.Errorf("runtime %q: %w", name, err)
			}

			continue
		}

		// only remotes from this point.

		rtsvc = sdkruntimesclient.New(sdkclient.Params{URL: addr, L: sl.Desugar()})

		if name != "*" {
			// specific from remote.

			if err := add(rtsvc, name, addr); err != nil {
				return nil, fmt.Errorf("runtime %q: %w", name, err)
			}

			continue
		}

		// all from remote.

		rrts, err := sdkruntimesclient.New(sdkclient.Params{URL: addr, L: sl.Desugar()}).List(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list runtimes at %s: %w", addr, err)
		}

		for _, rt := range rrts {
			if err := add(rtsvc, rt.Name().String(), addr); err != nil {
				return nil, err
			}
		}
	}

	return sdkruntimes.New(rts)
}
