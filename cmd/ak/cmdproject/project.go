package cmdproject

import (
	"github.com/urfave/cli/v2"

	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdproject/cmdcreate"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdproject/cmdget"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdproject/cmdsetup"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdproject/cmdupdate"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdproject/projecttools"
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
