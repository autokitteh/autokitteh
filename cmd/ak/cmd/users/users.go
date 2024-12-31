package users

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var usersCmd = common.StandardCommand(&cobra.Command{
	Use:   "users",
	Short: "Users: create, get",
	Args:  cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(usersCmd)
}

func init() {
	// Subcommands.
	usersCmd.AddCommand(createCmd)
	usersCmd.AddCommand(getCmd)
	usersCmd.AddCommand(getOrgsCmd)
	usersCmd.AddCommand(updateCmd)
}

func users() sdkservices.Users { return common.Client().Users() }
func orgs() sdkservices.Orgs   { return common.Client().Orgs() }
