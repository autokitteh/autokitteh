package connections

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flags shared by the "create" and "list" subcommands.
var integration, connectionToken string

var connectionsCmd = common.StandardCommand(&cobra.Command{
	Use:     "connections",
	Short:   "Connection management commands",
	Aliases: []string{"connection", "conn", "con", "c"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(connectionsCmd)
}

func init() {
	// Subcommands.
	connectionsCmd.AddCommand(createCmd)
	connectionsCmd.AddCommand(deleteCmd)
	connectionsCmd.AddCommand(getCmd)
	connectionsCmd.AddCommand(listCmd)
}

func connections() sdkservices.Connections {
	return common.Client().Connections()
}
