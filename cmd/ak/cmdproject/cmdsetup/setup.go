package cmdsetup

import (
	"github.com/urfave/cli/v2"

	T "gitlab.com/softkitteh/autokitteh/cmd/ak/clitools"
	P "gitlab.com/softkitteh/autokitteh/cmd/ak/cmdproject/projecttools"
)

var (
	Cmd = cli.Command{
		Name:    "setup",
		Aliases: []string{"s"},
		Action:  func(_ *cli.Context) error { return P.Projects().Setup(T.Context) },
	}
)
