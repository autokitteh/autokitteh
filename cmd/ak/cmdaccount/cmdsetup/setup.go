package cmdsetup

import (
	"github.com/urfave/cli/v2"

	T "gitlab.com/softkitteh/autokitteh/cmd/ak/clitools"
	A "gitlab.com/softkitteh/autokitteh/cmd/ak/cmdaccount/accounttools"
)

var (
	Cmd = cli.Command{
		Name:    "setup",
		Aliases: []string{"s"},
		Action:  func(_ *cli.Context) error { return A.Accounts().Setup(T.Context) },
	}
)
