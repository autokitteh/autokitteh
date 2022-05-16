package svc

import (
	"github.com/urfave/cli/v2"
)

type cliopts struct {
	app       []func(*cli.App)
	preaction []func(*cli.Context) error
	flags     []cli.Flag
}

type CLIOptFunc func(*cliopts)

func WithCLIApp(f func(*cli.App)) CLIOptFunc { return func(o *cliopts) { o.app = append(o.app, f) } }

func WithCLICommand(cmd *cli.Command) CLIOptFunc {
	return WithCLIApp(func(app *cli.App) {
		app.Commands = append(app.Commands, cmd)
	})
}

func WithCLIPreaction(f func(c *cli.Context) error) CLIOptFunc {
	return func(o *cliopts) { o.preaction = append(o.preaction, f) }
}

func WithCLIFlags(flags []cli.Flag) CLIOptFunc {
	return func(o *cliopts) { o.flags = append(o.flags, flags...) }
}
