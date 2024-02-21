package manifest

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var manifestCmd = common.StandardCommand(&cobra.Command{
	Use:     "manifest",
	Short:   "Manifest file commands",
	Aliases: []string{"man"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(manifestCmd)
}

func init() {
	// Subcommands.
	manifestCmd.AddCommand(applyCmd)
	manifestCmd.AddCommand(execCmd)
	manifestCmd.AddCommand(planCmd)
	manifestCmd.AddCommand(schemaCmd)
	manifestCmd.AddCommand(validateCmd)
}
