package integrations

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var integrationsCmd = common.StandardCommand(&cobra.Command{
	Use:     "integrations",
	Short:   "Integration management commands",
	Aliases: []string{"integration", "int", "i"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(integrationsCmd)
}

func init() {
	// Subcommands.
	// TODO: integrationsCmd.AddCommand(createCmd)
	// TODO: integrationsCmd.AddCommand(updateCmd)
	// TODO: integrationsCmd.AddCommand(deleteCmd)
	integrationsCmd.AddCommand(getCmd)
	integrationsCmd.AddCommand(listCmd)
}

func integrations() sdkservices.Integrations {
	return common.Client().Integrations()
}
