package cmdupdate

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	A "github.com/autokitteh/autokitteh/cmd/ak/cmdaccount/accounttools"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiaccount"
)

var (
	opts struct {
		name    string
		memo    cli.StringSlice
		disable bool
	}

	Cmd = cli.Command{
		Name:    "update",
		Aliases: []string{"u"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "name",
				Aliases:     []string{"n"},
				Destination: &opts.name,
			},
			&cli.StringSliceFlag{
				Name:        "memo",
				Aliases:     []string{"m"},
				Destination: &opts.memo,
			},
			&cli.BoolFlag{
				Name:        "disable",
				Aliases:     []string{"d"},
				Destination: &opts.disable,
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()

			if len(args) != 1 {
				return fmt.Errorf("single account id expected")
			}

			memo := make(map[string]string, len(opts.memo.Value()))
			for _, kv := range opts.memo.Value() {
				parts := strings.SplitN(kv, "=", 2)
				var v string
				if len(parts) > 1 {
					v = parts[1]
				}
				memo[parts[0]] = v
			}

			data := (&apiaccount.AccountData{}).SetName(opts.name).SetMemo(memo).SetEnabled(!opts.disable)

			if err := A.Accounts().Update(
				T.Context,
				apiaccount.AccountID(args[0]),
				data,
			); err != nil {
				return err
			}

			return nil
		},
	}
)
