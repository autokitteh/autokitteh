package users

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var getOrgsCmd = common.StandardCommand(&cobra.Command{
	Use:   "get-orgs [email or id] [--fail]",
	Short: "Get user orgs",
	Args:  cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		var (
			uid sdktypes.UserID
			err error
		)

		if len(args) > 0 {
			_, uid, err = r.User(ctx, args[0])
			err = common.AddNotFoundErrIfCond(err, uid.IsValid())
		}

		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "user"); err == nil {
			orgs, err := orgs().GetOrgsForUser(ctx, uid)
			if err != nil {
				return err
			}

			common.RenderKVIfV("orgs", orgs)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getOrgsCmd)
}
