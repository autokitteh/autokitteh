package auth

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var createTokenCmd = common.StandardCommand(&cobra.Command{
	Use:   "create-token",
	Short: "Create auth token",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		tok, err := auth().CreateToken(ctx)
		if err != nil {
			return err
		}

		common.Render(tok)

		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(createTokenCmd)
}
