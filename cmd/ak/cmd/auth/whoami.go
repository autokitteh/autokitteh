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
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "user", u.IsValid()); err != nil {
			return err
		}

		if u.IsValid() {
			common.RenderKV("user", u)
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(whoamiCmd)
}
