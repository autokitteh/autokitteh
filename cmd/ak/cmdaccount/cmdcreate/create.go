package cmdcreate

import (
	"strings"

	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	A "github.com/autokitteh/autokitteh/cmd/ak/cmdaccount/accounttools"
	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"github.com/autokitteh/autokitteh/sdk/api/apiaccount"
)

var (
	opts struct {
		name string
		memo cli.StringSlice
	}

	Cmd = cli.Command{
		Name:    "create",
		Aliases: []string{"c"},
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
		},
		Action: func(c *cli.Context) error {
			memo := make(map[string]string, len(opts.memo.Value()))
			for _, kv := range opts.memo.Value() {
				parts := strings.SplitN(kv, "=", 2)
				var v string
				if len(parts) > 1 {
					v = parts[1]
				}
				memo[parts[0]] = v
			}

			data := (&apiaccount.AccountData{}).SetName(opts.name).SetMemo(memo)

			id, err := A.Accounts().Create(
				T.Context,
				accountsstore.AutoAccountID,
				data,
			)

			if err != nil {
				return err
			}

			T.Show(
				map[string]string{
					"id": id.String(),
				},
			)

			return nil
		},
	}
)
