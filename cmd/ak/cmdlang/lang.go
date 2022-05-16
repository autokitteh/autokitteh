package cmdlang

import (
	"github.com/urfave/cli/v2"

	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/langtools"

	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/cmdcall"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/cmdcat"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/cmdcompile"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/cmddeps"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/cmdlist"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/cmdrun"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdlang/cmdruns"

	_ "github.com/autokitteh/autokitteh/internal/pkg/lang/langall"
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
