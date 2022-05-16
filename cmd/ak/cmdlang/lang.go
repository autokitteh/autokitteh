package cmdlang

import (
	"github.com/urfave/cli/v2"

	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/langtools"

	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/cmdcall"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/cmdcat"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/cmdcompile"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/cmddeps"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/cmdlist"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/cmdrun"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang/cmdruns"

	_ "gitlab.com/softkitteh/autokitteh/internal/pkg/lang/langall"
)

var (
	Cmd = cli.Command{
		Name:    "lang",
		Usage:   "languages",
		Aliases: []string{"l"},
		Subcommands: []*cli.Command{
			&cmdcall.Cmd,
			&cmdcat.Cmd,
			&cmdcompile.Cmd,
			&cmddeps.Cmd,
			&cmdlist.Cmd,
			&cmdrun.Cmd,
			&cmdruns.Cmd,
		},
		Before: func(c *cli.Context) error {
			if err := langtools.Init(); err != nil {
				return err
			}

			return nil
		},
	}
)
