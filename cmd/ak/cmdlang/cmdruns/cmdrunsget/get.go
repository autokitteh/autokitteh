package cmdrunsget

import (
	"fmt"

	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/langtools"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langrun"
)

var (
	Cmd = cli.Command{
		Name:    "get",
		Aliases: []string{"g"},
		Action: func(c *cli.Context) error {
			ctx := T.Context

			args := c.Args().Slice()

			if len(args) == 1 && args[0] == "all" {
				args = nil

				rs, err := L.Runs().List(ctx)
				if err != nil {
					return fmt.Errorf("list: %w", err)
				}

				for _, kvs := range rs {
					for k := range kvs {
						args = append(args, string(k))
					}
				}

				T.L().Info("getting all runs", "ids", args)
			}

			rs := make(map[string]interface{})

			for _, arg := range args {
				r, err := L.Runs().Get(ctx, langrun.RunID(arg))
				if err != nil {
					return fmt.Errorf("get %s: %w", arg, err)
				}

				s, err := r.Summary(ctx)
				if err != nil {
					return fmt.Errorf("summary %s: %w", arg, err)
				}

				rs[arg] = s
			}

			T.Show(rs)

			return nil
		},
	}
)
