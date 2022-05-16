package cmdproject

import (
	"github.com/urfave/cli/v2"

	"github.com/autokitteh/autokitteh/cmd/ak/cmdproject/cmdcreate"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdproject/cmdget"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdproject/cmdsetup"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdproject/cmdupdate"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdproject/projecttools"
)

var (
	flags struct {
		cfg string
	}

	Cmd = cli.Command{
		Name:    "project",
		Usage:   "projects",
		Aliases: []string{"p"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c", "cfg"},
				Destination: &flags.cfg,
			},
		},
		Subcommands: []*cli.Command{
			&cmdcreate.Cmd,
			&cmdget.Cmd,
			&cmdsetup.Cmd,
			&cmdupdate.Cmd,
		},
		Before: func(c *cli.Context) error {
			if err := projecttools.Init(flags.cfg); err != nil {
				return err
			}

			return nil
		},
	}
)
