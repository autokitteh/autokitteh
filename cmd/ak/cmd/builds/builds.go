package builds

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

const (
	defaultOutput = "build.akb"
)

var buildsCmd = common.StandardCommand(&cobra.Command{
	Use:     "builds",
	Short:   "Build management commands",
	Aliases: []string{"build", "b"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(buildsCmd)
}

func init() {
	// Subcommands.
	buildsCmd.AddCommand(deleteCmd)
	buildsCmd.AddCommand(describeCmd)
	buildsCmd.AddCommand(downloadCmd)
	buildsCmd.AddCommand(getCmd)
	buildsCmd.AddCommand(listCmd)
	buildsCmd.AddCommand(uploadCmd)
}

func builds() sdkservices.Builds {
	return common.Client().Builds()
}
