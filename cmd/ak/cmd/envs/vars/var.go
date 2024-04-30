package vars

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flags shared by all the subcommands.
var env, project string

var varCmd = common.StandardCommand(&cobra.Command{
	Use:   "var",
	Short: "Environment variable subcommands: set, get, reveal, remove",
	Args:  cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(varCmd)
}

func init() {
	// Flags shared by all subcommands.
	// We don't define them as single persistent flags here
	// for aesthetic conformance with flags in other "env" sibling commands.
	getCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	removeCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	revealCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	setCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")

	getCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	removeCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	revealCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	setCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")

	// Subcommands.
	varCmd.AddCommand(getCmd)
	varCmd.AddCommand(removeCmd)
	varCmd.AddCommand(revealCmd)
	varCmd.AddCommand(setCmd)
}

func envs() sdkservices.Envs {
	return common.Client().Envs()
}
