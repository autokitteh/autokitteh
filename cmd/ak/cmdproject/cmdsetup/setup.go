package cmdsetup

import (
	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	P "github.com/autokitteh/autokitteh/cmd/ak/cmdproject/projecttools"
)

var (
	Cmd = cli.Command{
		Name:    "setup",
		Aliases: []string{"s"},
		Action:  func(_ *cli.Context) error { return P.Projects().Setup(T.Context) },
	}
)
