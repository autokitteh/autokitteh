package aksvc

import (
	"github.com/urfave/cli/v2"

	"github.com/autokitteh/svc"
)

// See [# InitPaths #] in config.go
var initPaths cli.StringSlice

type Version = svc.Version

func Run(version *Version) {
	svc.SetVersion(version)

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
