package cmdparseargs

import (
	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
)

var (
	opts struct {
		unwrapped bool
	}

	Cmd = cli.Command{
		Name:   "parse-args",
		Hidden: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "unwrapped",
				Aliases:     []string{"u"},
				Destination: &opts.unwrapped,
			},
		},
		Action: func(c *cli.Context) error {
			l, m, err := T.ParseValuesArgs(c.Args().Slice(), opts.unwrapped)

			T.Show(struct {
				L []*apivalues.Value          `json:"args"`
				M map[string]*apivalues.Value `json:"kwargs"`
			}{
				L: l,
				M: m,
			})

			return err
		},
	}
)
