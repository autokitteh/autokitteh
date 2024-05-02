package envs

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flag shared by all the subcommands.
var project string

var envCmd = common.StandardCommand(&cobra.Command{
	Use:   "env",
	Short: "Execution environments: create, get, list",
	Args:  cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(envCmd)
}

func init() {
	// Flag shared by all subcommands.
	// We don't define it as a single persistent flag here
	// because then we wouldn't be able to mark it as required.
	createCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	kittehs.Must0(createCmd.MarkFlagRequired("project"))

	getCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	listCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")

	// Subcommands.
	envCmd.AddCommand(createCmd)
	envCmd.AddCommand(getCmd)
	envCmd.AddCommand(listCmd)
}

func envs() sdkservices.Envs {
	return common.Client().Envs()
}
