package cmdget

import (
	"fmt"

	"github.com/urfave/cli/v2"

	T "gitlab.com/softkitteh/autokitteh/cmd/ak/clitools"
	P "gitlab.com/softkitteh/autokitteh/cmd/ak/cmdproject/projecttools"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiproject"
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
