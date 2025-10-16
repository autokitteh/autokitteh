package temporal

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var serverCmd = common.StandardCommand(&cobra.Command{
	Use:   "temporal",
	Short: "Temporal utilities",
	Args:  cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(serverCmd)
}

func init() {
	// Subcommands.
	serverCmd.AddCommand(downloadCmd)
}
