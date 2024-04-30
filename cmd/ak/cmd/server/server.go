package server

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

// Flag shared by all subcommands.
var mode string

var serverCmd = common.StandardCommand(&cobra.Command{
	Use:     "server",
	Short:   "Local server and storage",
	Aliases: []string{"srv"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(serverCmd)
}

func Remove(parentCmd *cobra.Command) {
	parentCmd.RemoveCommand(serverCmd)
}

func init() {
	// Flag shared by all subcommands.
	serverCmd.PersistentFlags().StringVarP(&mode, "mode", "m", "", "run mode")

	// Subcommands.
	serverCmd.AddCommand(migrateCmd)
}
