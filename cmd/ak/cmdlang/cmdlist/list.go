package cmdlist

import (
	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	L "github.com/autokitteh/autokitteh/cmd/ak/cmdlang/langtools"
)

var (
	Cmd = cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Action: func(c *cli.Context) error {
			T.Show(map[string]interface{}{"langs": L.Catalog().List()})
			return nil
		},
	}
)
