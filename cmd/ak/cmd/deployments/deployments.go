package deployments

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flags shared by the "create" and "list" subcommands.
var buildID, env string

var deploymentsCmd = common.StandardCommand(&cobra.Command{
	Use:     "deployments",
	Short:   "Build deployment management commands",
	Aliases: []string{"deployment", "deploy", "dep", "d"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(deploymentsCmd)
}

func init() {
	// Subcommands.
	deploymentsCmd.AddCommand(createCmd)
	deploymentsCmd.AddCommand(activateCmd)
	deploymentsCmd.AddCommand(drainCmd)
	deploymentsCmd.AddCommand(deactivateCmd)
	deploymentsCmd.AddCommand(getCmd)
	deploymentsCmd.AddCommand(listCmd)
}

func deployments() sdkservices.Deployments {
	return common.Client().Deployments()
}
