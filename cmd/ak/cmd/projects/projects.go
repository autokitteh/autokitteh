package projects

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var projectsCmd = common.StandardCommand(&cobra.Command{
	Use:     "projects",
	Short:   "Project management commands",
	Aliases: []string{"project", "proj", "p"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(projectsCmd)
}

func init() {
	// Subcommands.
	projectsCmd.AddCommand(buildCmd)
	projectsCmd.AddCommand(createCmd)
	projectsCmd.AddCommand(deployCmd)
	projectsCmd.AddCommand(downloadCmd)
	projectsCmd.AddCommand(getCmd)
	projectsCmd.AddCommand(listCmd)
}

func projects() sdkservices.Projects {
	return common.Client().Projects()
}
