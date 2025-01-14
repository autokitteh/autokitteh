package users

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var getIDCmd = common.StandardCommand(&cobra.Command{
	Use:   "get-id <email> [--fail]",
	Short: "Get user ID",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		uid, err := users().GetID(ctx, args[0])
		err = common.AddNotFoundErrIfCond(err, uid.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "user_id"); err == nil {
			common.RenderKVIfV("user_id", uid)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getIDCmd)
}
