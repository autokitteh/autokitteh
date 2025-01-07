package users

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:   "get <email or id> [--fail]",
	Short: "Get user details",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		u, _, err := r.User(ctx, args[0])
		err = common.AddNotFoundErrIfCond(err, u.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "user"); err == nil {
			common.RenderKVIfV("user", u)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
