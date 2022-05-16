package cmdruns

import (
	"github.com/urfave/cli/v2"

	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/cmdruns/cmdrunscancel"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/cmdruns/cmdrunsdiscard"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/cmdruns/cmdrunsget"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/cmdruns/cmdrunslist"
)

var (
	Cmd = cli.Command{
		Name:    "runs",
		Usage:   "runs",
		Aliases: []string{"rs"},
		Subcommands: []*cli.Command{
			&cmdrunscancel.Cmd,
			&cmdrunsget.Cmd,
			&cmdrunslist.Cmd,
			&cmdrunsdiscard.Cmd,
		},
	}
)
