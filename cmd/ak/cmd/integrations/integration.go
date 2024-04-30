package integrations

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var integrationCmd = common.StandardCommand(&cobra.Command{
	Use:     "integration",
	Short:   "Integrations: create, get, list, update, delete",
	Aliases: []string{"int"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(integrationCmd)
}

func init() {
	// Subcommands.
	// TODO: integrationCmd.AddCommand(createCmd)
	// TODO: integrationCmd.AddCommand(deleteCmd)
	integrationCmd.AddCommand(getCmd)
	integrationCmd.AddCommand(listCmd)
	// TODO: integrationCmd.AddCommand(updateCmd)
}

func integrations() sdkservices.Integrations {
	return common.Client().Integrations()
}
