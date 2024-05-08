package auth

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var setTokenCmd = common.StandardCommand(&cobra.Command{
	Use:   "set-token <token>",
	Short: "Set authentication token",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		return common.StoreToken(args[0])
	},
})
