package cmdaccount

import (
	"github.com/urfave/cli/v2"

	"github.com/autokitteh/autokitteh/cmd/ak/cmdaccount/accounttools"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdaccount/cmdcreate"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdaccount/cmdget"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdaccount/cmdsetup"
	"github.com/autokitteh/autokitteh/cmd/ak/cmdaccount/cmdupdate"
)

var (
	flags struct {
		cfg string
	}

	Cmd = cli.Command{
		Name:    "account",
		Usage:   "accounts",
		Aliases: []string{"a"},
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
			if err := accounttools.Init(flags.cfg); err != nil {
				return err
			}

			return nil
		},
	}
)
