package experimental

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var experimentalCmd = common.StandardCommand(&cobra.Command{
	Use:     "experimental",
	Short:   "Unofficial or internal commands",
	Aliases: []string{"x"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(experimentalCmd)
}

func init() {
	// Subcommands.
	experimentalCmd.AddCommand(downCmd)
}
