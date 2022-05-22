package cmdget

import (
	"fmt"

	"github.com/urfave/cli/v2"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	A "github.com/autokitteh/autokitteh/cmd/ak/cmdaccount/accounttools"
	"github.com/autokitteh/autokitteh/sdk/api/apiaccount"
)

var (
	Cmd = cli.Command{
		Name:    "get",
		Aliases: []string{"g"},
		Action: func(c *cli.Context) error {
			as := make(map[string]interface{})

			for _, arg := range c.Args().Slice() {
				a, err := A.Accounts().Get(
					T.Context,
					apiaccount.AccountID(arg),
				)

				if err != nil {
					return fmt.Errorf("get %s: %w", arg, err)
				}

				as[arg] = a
			}

			T.Show(as)

			return nil
		},
	}
)
