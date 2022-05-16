package cmdrunslist

import (
	"fmt"

	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	L "github.com/autokitteh/autokitteh/cmd/ak/cmdlang/langtools"
)

var (
	Cmd = cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Action: func(c *cli.Context) error {
			rs, err := L.Runs().List(T.Context)
			if err != nil {
				return fmt.Errorf("list: %w", err)
			}

			T.Show(map[string]interface{}{"runs": rs})

			return nil
		},
	}
)
