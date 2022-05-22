package cmdrunsdiscard

import (
	"fmt"

	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/langtools"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langrun"
)

var (
	Cmd = cli.Command{
		Name:    "discard",
		Aliases: []string{"d"},
		Action: func(c *cli.Context) error {
			ctx := T.Context

			args := c.Args().Slice()

			if len(args) == 1 && args[0] == "all" {
				args = nil

				rs, err := L.Runs().List(ctx)
				if err != nil {
					return fmt.Errorf("list: %w", err)
				}

				for _, state := range []string{"completed", "error", "canceled"} {
					if runnings, ok := rs[state]; ok {
						for k := range runnings {
							args = append(args, string(k))
						}
					}
				}

				T.L().Info("discarding all runs", "ids", args)
			}

			for _, arg := range args {
				run, err := L.Runs().Get(ctx, langrun.RunID(arg))
				if err != nil {
					return fmt.Errorf("get %s: %w", arg, err)
				}

				if run == nil {
					return fmt.Errorf("get %s: not found", arg)
				}

				if err := run.Discard(ctx); err != nil {
					return fmt.Errorf("discard %s: %w", arg, err)
				}
			}

			return nil
		},
	}
)
