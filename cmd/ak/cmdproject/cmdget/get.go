package cmdget

import (
	"fmt"

	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	P "github.com/autokitteh/autokitteh/cmd/ak/cmdproject/projecttools"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
)

var (
	Cmd = cli.Command{
		Name:    "get",
		Aliases: []string{"g"},
		Action: func(c *cli.Context) error {
			as := make(map[string]interface{})

			for _, arg := range c.Args().Slice() {
				a, err := P.Projects().Get(
					T.Context,
					apiproject.ProjectID(arg),
				)

				if err != nil {
					return fmt.Errorf("get %s: %w", arg, err)
				}

				as[arg] = a
			}

			T.Show(as)

			return nil
		},
	}
)
