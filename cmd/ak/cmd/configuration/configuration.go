package configuration

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var configurationCmd = common.StandardCommand(&cobra.Command{
	Use:     "configuration",
	Short:   "Configuration management commands",
	Aliases: []string{"config", "conf", "cfg"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(configurationCmd)
}

func init() {
	// Subcommands.
	configurationCmd.AddCommand(listCmd)
	configurationCmd.AddCommand(setCmd)
	configurationCmd.AddCommand(whereCmd)
}
