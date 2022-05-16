package cmdrun

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/urfave/cli/v2"

	T "gitlab.com/softkitteh/autokitteh/cmd/ak/clitools"
	L "gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/langtools"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang/langtools"
)

var (
	opts struct {
		unwrap   bool
		scope    string
		predecls cli.StringSlice
	}

	Cmd = cli.Command{
		Name:    "run",
		Aliases: []string{"r"},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "unwrap",
				Aliases:     []string{"u"},
				Usage:       "unwrap result values",
				Destination: &opts.unwrap,
			},
			&cli.StringFlag{
				Name:        "scope",
				Aliases:     []string{"s"},
				Usage:       "run scope",
				Destination: &opts.scope,
			},
			&cli.StringSliceFlag{
				Name:        "predecl",
				Aliases:     []string{"p"},
				Destination: &opts.predecls,
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()

			if len(args) < 1 {
				return fmt.Errorf("at least a single ak program file path is expected")
			}

			l, predecls, err := T.ParseValuesArgs(opts.predecls.Value(), !opts.unwrap)
			if err != nil {
				return fmt.Errorf("predecls: %w", err)
			}
			if len(l) != 0 {
				return fmt.Errorf("all predecls must be in k=v format")
			}

			mods := make([]*apiprogram.Module, len(args))
			for i, arg := range args {
				var err error
				if mods[i], err = L.Load(arg); err != nil {
					return fmt.Errorf("load %q: %w", arg, err)
				}
			}

			ctx := T.Context

			var gs map[string]*apivalues.Value

			env := lang.RunEnv{
				Scope:    opts.scope,
				Print:    func(s string) { fmt.Println(s) },
				Predecls: predecls,
			}

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt)

			go func() {
				sig := <-ch

				T.Warnw("got signal", "sig", sig)

				cancel()
			}()

			gs, _, err = langtools.RunModules(ctx, L.Catalog(), &env, mods)

			var vs interface{}

			if opts.unwrap {
				vs = apivalues.UnwrapValuesMap(gs, apivalues.WithUnwrapJSONSafe())
			} else {
				vs = gs
			}

			T.ShowStderr(
				struct {
					Values interface{} `json:"values"`
				}{
					Values: vs,
				},
			)

			return err
		},
	}
)
