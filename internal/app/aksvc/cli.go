package aksvc

import (
	"github.com/urfave/cli/v2"
	"gitlab.com/softkitteh/autokitteh/pkg/svc"
)

// See [# InitPaths #] in config.go
var initPaths cli.StringSlice

func Run() {
	svc.RunCLI(
		"",
		append(
			SvcOpts,
			svc.WithCLIOptions(
				svc.WithCLIFlags(
					[]cli.Flag{
						&cli.StringSliceFlag{
							Name:        "initpath",
							Aliases:     []string{"i"},
							Destination: &initPaths,
							Usage:       "list of paths to initialization manifests",
						},
					},
				),
			),
		)...,
	)
}
