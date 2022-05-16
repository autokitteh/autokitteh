package cmdcall

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/urfave/cli/v2"

	T "gitlab.com/softkitteh/autokitteh/cmd/ak/clitools"
	L "gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/langtools"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang/langtools"
)

var (
	opts struct {
		unwrap bool
		scope  string
	}

	Cmd = cli.Command{
		Name: "call",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "unwrap",
				Aliases:     []string{"u"},
				Usage:       "unwrap results",
				Destination: &opts.unwrap,
			},
			&cli.StringFlag{
				Name:        "scope",
				Aliases:     []string{"s"},
				Usage:       "call scope",
				Destination: &opts.scope,
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()

			if len(args) < 1 {
				return fmt.Errorf("function value expected")
			}

			fnv, err := apivalues.ParseFunctionValueString(args[0])
			if err != nil {
				return fmt.Errorf("invalid function: %w", err)
			}

			fn, err := apivalues.NewValue(*fnv)
			if err != nil {
				return fmt.Errorf("invalid function value: %w", err)
			}

			l, m, err := T.ParseValuesArgs(args[1:], !opts.unwrap)
			if err != nil {
				return fmt.Errorf("args: %w", err)
			}

			T.L().Debug("args", "l", l, "m", m)

			if opts.scope == "" {
				opts.scope = fnv.Scope
			}

			ctx := T.Context

			env := lang.RunEnv{
				Scope: opts.scope,
				Print: func(s string) { fmt.Println(s) },
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

			_, g, _, err := langtools.CallFunction(ctx, L.Catalog(), &env, fn, l, m)

			var v interface{}

			if g != nil {
				if opts.unwrap {
					v = apivalues.Unwrap(g.Get(), apivalues.WithUnwrapJSONSafe())
				} else {
					v = g
				}
			}

			T.ShowStderr(
				struct {
					Values interface{} `json:"value"`
				}{
					Values: v,
				},
			)

			return err
		},
	}
)
