package cmdresettestidgen

import (
	"github.com/urfave/cli/v2"

	"gitlab.com/softkitteh/autokitteh/pkg/idgen"
)

var (
	opts struct {
		n uint64
	}

	Cmd = cli.Command{
		Name:  "reset-test-idgen",
		Usage: "for testing only - reset id generation",
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:        "n",
				Destination: &opts.n,
			},
		},
		Action: func(c *cli.Context) error {
			idgen.New = idgen.NewSequentialPerPrefix(opts.n)
			return nil
		},
	}
)
