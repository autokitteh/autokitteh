package cmdruns

import (
	"github.com/urfave/cli/v2"

	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/cmdruns/cmdrunscancel"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/cmdruns/cmdrunsdiscard"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/cmdruns/cmdrunsget"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/cmdruns/cmdrunslist"
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
