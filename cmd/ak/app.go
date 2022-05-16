package main

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"gitlab.com/softkitteh/autokitteh/cmd/ak/clitools"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdaccount"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdbumptestidgen"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdlang"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdparseargs"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdproject"
	"gitlab.com/softkitteh/autokitteh/cmd/ak/cmdresettestidgen"

	"gitlab.com/softkitteh/autokitteh/internal/app/aksvc"
	"gitlab.com/softkitteh/autokitteh/pkg/idgen"
	L "gitlab.com/softkitteh/autokitteh/pkg/l"
	"gitlab.com/softkitteh/autokitteh/pkg/svc"
)

var (
	flags struct {
		yaml       bool
		indentJSON bool
		logLevel   string
		tToIntr    time.Duration
		tToCancel  time.Duration
		testIDGen  bool
		addr       string
	}

	newApp func() *cli.App

	first = true
)

func init() {
	newApp = func() *cli.App {
		return &cli.App{
			Name:  "ak",
			Usage: "autokitteh cli",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "addr",
					Aliases:     []string{"a"},
					Destination: &flags.addr,
					Value:       "builtin",
				},
				&cli.BoolFlag{
					Name:        "yaml",
					Aliases:     []string{"y"},
					Destination: &flags.yaml,
					Usage:       "use YAML for output data",
				},
				&cli.BoolFlag{
					Name:        "test-idgen",
					Destination: &flags.testIDGen,
					Usage:       "for testing only - make id generation predictable",
				},
				&cli.BoolFlag{
					Name:        "indent-json",
					Aliases:     []string{"ij"},
					Destination: &flags.indentJSON,
					Usage:       "if JSON output, indent it",
				},
				&cli.StringFlag{
					Name:        "log-level",
					Aliases:     []string{"ll"},
					Destination: &flags.logLevel,
					DefaultText: "warn",
				},
				&cli.DurationFlag{
					Name:        "time-to-interrupt",
					Aliases:     []string{"tti"},
					Destination: &flags.tToIntr,
				},
				&cli.DurationFlag{
					Name:        "time-to-cancel",
					Aliases:     []string{"ttc"},
					Destination: &flags.tToCancel,
				},
			},
			Commands: []*cli.Command{
				&cmdSh,
				&cmdaccount.Cmd,
				&cmdbumptestidgen.Cmd,
				&cmdlang.Cmd,
				&cmdparseargs.Cmd,
				&cmdproject.Cmd,
				&cmdresettestidgen.Cmd,
				svc.CLICmd("svc", append(aksvc.SvcOpts, svc.WithLogger(func() L.L { return clitools.L() }))...),
			},
			Before: func(c *cli.Context) error {
				if !first {
					return nil
				}

				first = false

				if flags.testIDGen {
					idgen.New = idgen.NewSequentialPerPrefix(0)
				}

				clitools.Settings.Yaml = flags.yaml
				clitools.Settings.IndentJSON = flags.indentJSON
				clitools.Settings.LogLevel = flags.logLevel

				if err := clitools.Init(flags.addr); err != nil {
					return fmt.Errorf("init cli: %w", err)
				}

				if t := flags.tToIntr; t != 0 {
					go func() {
						time.Sleep(t)
						_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
					}()
				}

				if t := flags.tToCancel; t != 0 {
					var cancel func()
					clitools.Context, cancel = context.WithCancel(clitools.Context)

					go func() {
						time.Sleep(t)
						cancel()
					}()
				}

				return nil
			},
		}
	}
}
