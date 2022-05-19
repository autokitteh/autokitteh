package svc

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func FlagsAndAction(optsargs ...OptFunc) ([]cli.Flag, func(c *cli.Context) error, *cliopts) {
	var svcopts opts

	for _, opt := range optsargs {
		opt(&svcopts)
	}

	var cliopts cliopts
	for _, opt := range svcopts.cli {
		opt(&cliopts)
	}

	var (
		flags                             Flags
		enables, disables, onlys, excepts cli.StringSlice
		ver, bg                           bool
	)

	cliFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Destination: &flags.ConfigPath,
			Usage:       "use config file",
		},
		&cli.StringSliceFlag{
			Name:        "only",
			Destination: &onlys,
			Usage:       "enable only these modules",
		},
		&cli.StringSliceFlag{
			Name:        "except",
			Destination: &excepts,
			Usage:       "disable only these modules",
		},
		&cli.StringSliceFlag{
			Name:        "enable",
			Destination: &enables,
			Usage:       "modules to enable",
		},
		&cli.StringSliceFlag{
			Name:        "disable",
			Destination: &disables,
			Usage:       "modules to disable",
		},
		&cli.BoolFlag{
			Name:        "background",
			Aliases:     []string{"bg"},
			Destination: &bg,
			Hidden:      true,
			Usage:       "run in a separate go routine",
		},
		&cli.BoolFlag{
			Name:        "setup",
			Destination: &flags.Setup,
			Usage:       "run setup phase",
		},
		&cli.BoolFlag{
			Name:        "help-config",
			Destination: &flags.HelpConfig,
			Usage:       "describe accepted environment variables and exit",
		},
		&cli.BoolFlag{
			Name:        "exit-before-start",
			Destination: &flags.ExitBeforeStart,
			Usage:       "exit before start phase",
		},
		&cli.BoolFlag{
			Name:        "print-config",
			Destination: &flags.PrintConfig,
			Usage:       "print configuration",
		},
	}

	if GetVersion() != nil {
		cliFlags = append(cliFlags, &cli.BoolFlag{
			Name:        "version",
			Destination: &ver,
			Usage:       "print version and exit",
		})
	}

	cliAction := func(c *cli.Context) error {
		if ver {
			fmt.Println(GetVersion().String())
			return nil
		}

		for _, a := range cliopts.preaction {
			if err := a(c); err != nil {
				return err
			}
		}

		flags.Enables = enables.Value()
		flags.Disables = disables.Value()
		flags.Onlys = onlys.Value()
		flags.Excepts = excepts.Value()

		run := func() { Run(append(optsargs, WithFlags(&flags))...) }

		if bg {
			go run()
		} else {
			run()
		}

		return nil
	}

	return append(cliFlags, cliopts.flags...), cliAction, &cliopts
}

const usage = "autokitteh service"

func CLICmd(name string, opts ...OptFunc) *cli.Command {
	flags, action, _ := FlagsAndAction(opts...)

	return &cli.Command{Name: name, Usage: usage, Flags: flags, Action: action}
}

func RunCLI(name string, opts ...OptFunc) {
	flags, action, cliopts := FlagsAndAction(opts...)

	if name == "" {
		name = DefaultServiceName
	}

	app := &cli.App{Name: name, Usage: usage, Flags: flags, Action: action}

	for _, f := range cliopts.app {
		f(app)
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
