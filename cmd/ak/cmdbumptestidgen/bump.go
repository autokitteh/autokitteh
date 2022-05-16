package cmdbumptestidgen

import (
	"github.com/urfave/cli/v2"

	"gitlab.com/softkitteh/autokitteh/pkg/idgen"
)

var (
	Cmd = cli.Command{
		Name:  "bump-test-idgen",
		Usage: "for testing only - bump id",
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()
			for _, arg := range args {
				idgen.New(arg)
			}
			return nil
		},
	}
)
