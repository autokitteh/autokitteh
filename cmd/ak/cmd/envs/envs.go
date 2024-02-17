package envs

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/envs/vars"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flag shared by all the subcommands.
var project string

var envsCmd = common.StandardCommand(&cobra.Command{
	Use:     "envs",
	Short:   "Execution environment management commands",
	Aliases: []string{"env", "en"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(envsCmd)
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
	envsCmd.AddCommand(createCmd)
	envsCmd.AddCommand(getCmd)
	envsCmd.AddCommand(listCmd)

	vars.AddSubcommands(envsCmd)
}

func envs() sdkservices.Envs {
	return common.Client().Envs()
}
