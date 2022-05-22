package cmddeps

import (
	"fmt"

	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/langtools"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
)

var (
	Cmd = cli.Command{
		Name:    "deps",
		Aliases: []string{"dep"},
		Action: func(c *cli.Context) error {
			ctx := T.Context

			args := c.Args().Slice()

			for _, arg := range args {
				mod, err := L.Load(arg)
				if err != nil {
					return fmt.Errorf("load: %w", err)
				}

				l, err := langtools.GetModuleDependencies(ctx, L.Catalog(), mod)
				if err != nil {
					return fmt.Errorf("%s: %w", arg, err)
				}
				T.Show(map[string]interface{}{arg: l})
			}

			return nil
		},
	}
)
