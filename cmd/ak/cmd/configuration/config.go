package configuration

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var configCmd = common.StandardCommand(&cobra.Command{
	Use:     "config",
	Short:   "Persistent configurations: list, set, where",
	Aliases: []string{"cfg"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(configCmd)
}

func init() {
	// Subcommands.
	configCmd.AddCommand(listCmd)
	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(whereCmd)
}
