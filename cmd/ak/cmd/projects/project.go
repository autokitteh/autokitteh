package projects

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flags shared by the "build" and "deploy" subcommands.
var (
	dirPaths, filePaths []string
)

var projectCmd = common.StandardCommand(&cobra.Command{
	Use:     "project",
	Short:   "Projects: create, get, list, build, download, deploy, delete",
	Aliases: []string{"prj"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(projectCmd)
}

func init() {
	// Subcommands.
	projectCmd.AddCommand(buildCmd)
	projectCmd.AddCommand(createCmd)
	projectCmd.AddCommand(deleteCmd)
	projectCmd.AddCommand(deployCmd)
	projectCmd.AddCommand(downloadCmd)
	projectCmd.AddCommand(getCmd)
	projectCmd.AddCommand(listCmd)
}

func projects() sdkservices.Projects {
	return common.Client().Projects()
}
