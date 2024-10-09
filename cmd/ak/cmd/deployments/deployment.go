package deployments

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flags shared by the "create" and "list" subcommands.
var buildID, env string

var deploymentCmd = common.StandardCommand(&cobra.Command{
	Use:     "deployment",
	Short:   "Build deployments: create, (de)activate, get, list, drain, delete",
	Aliases: []string{"dep"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(deploymentCmd)
}

func init() {
	// Subcommands.
	deploymentCmd.AddCommand(activateCmd)
	deploymentCmd.AddCommand(createCmd)
	deploymentCmd.AddCommand(deactivateCmd)
	deploymentCmd.AddCommand(deleteCmd)
	deploymentCmd.AddCommand(getCmd)
	deploymentCmd.AddCommand(listCmd)
}

func deployments() sdkservices.Deployments {
	return common.Client().Deployments()
}
