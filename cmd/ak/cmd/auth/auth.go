package auth

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var authCmd = common.StandardCommand(&cobra.Command{
	Use:   "auth",
	Short: "Authentication: create-token, whoami",
	Args:  cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(authCmd)
}

func init() {
	// Subcommands.
	authCmd.AddCommand(whoamiCmd)
	authCmd.AddCommand(createTokenCmd)
}

func auth() sdkservices.Auth { return common.Client().Auth() }
