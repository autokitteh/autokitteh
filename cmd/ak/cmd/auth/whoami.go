package auth

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var whoamiCmd = common.StandardCommand(&cobra.Command{
	Use:   "whoami",
	Short: "Who am I",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		u, err := auth().WhoAmI(ctx)
		err = common.AddNotFoundErrIfCond(err, u.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "user"); err == nil {
			common.RenderKV("user", u)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(whoamiCmd)
}
