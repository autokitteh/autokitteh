package builds

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

const (
	defaultOutput = "build.akb"
)

var buildCmd = common.StandardCommand(&cobra.Command{
	Use:     "build",
	Short:   "Project resource builds: upload, download, get, list, describe, delete",
	Aliases: []string{"bld"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(buildCmd)
}

func init() {
	// Subcommands.
	buildCmd.AddCommand(deleteCmd)
	buildCmd.AddCommand(describeCmd)
	buildCmd.AddCommand(downloadCmd)
	buildCmd.AddCommand(getCmd)
	buildCmd.AddCommand(listCmd)
	buildCmd.AddCommand(uploadCmd)
}

func builds() sdkservices.Builds {
	return common.Client().Builds()
}
