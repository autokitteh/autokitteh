package cmdrunscancel

import (
	"fmt"

	"github.com/urfave/cli/v2"

	T "gitlab.com/softkitteh/autokitteh/cmd/ak/clitools"
	L "gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/langtools"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang/langrun"
)

var (
	opts struct {
		reason string
	}

	Cmd = cli.Command{
		Name:    "cancel",
		Aliases: []string{"c"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "reason",
				Aliases:     []string{"r"},
				Usage:       "cancel reason",
				Destination: &opts.reason,
			},
		},
		Action: func(c *cli.Context) error {
			ctx := T.Context

			args := c.Args().Slice()

			if len(args) == 1 && args[0] == "all" {
				args = nil

				rs, err := L.Runs().List(ctx)
				if err != nil {
					return fmt.Errorf("list: %w", err)
				}

				if runnings, ok := rs["running"]; ok {
					for k := range runnings {
						args = append(args, string(k))
					}
				}

				T.L().Info("cancelling all runs", "ids", args)
			}

			for _, arg := range args {
				run, err := L.Runs().Get(ctx, langrun.RunID(arg))
				if err != nil {
					return fmt.Errorf("get %s: %w", arg, err)
				}

				if run == nil {
					return fmt.Errorf("get %s: not found", arg)
				}

				if err := run.Cancel(ctx, opts.reason); err != nil {
					return fmt.Errorf("cancel %s: %w", arg, err)
				}
			}

			return nil
		},
	}
)
