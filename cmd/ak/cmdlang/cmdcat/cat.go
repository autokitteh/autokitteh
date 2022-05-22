package cmdcat

import (
	"fmt"

	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/langtools"
)

var (
	Cmd = cli.Command{
		Name: "cat",
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()

			for _, arg := range args {
				mod, err := L.Load(arg)
				if err != nil {
					return fmt.Errorf("load %q: %w", arg, err)
				}

				T.Show(
					map[string]interface{}{arg: mod},
				)
			}

			return nil
		},
	}
)
