package connections

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flags shared by the "create" and "list" subcommands.
var integration string

var connectionCmd = common.StandardCommand(&cobra.Command{
	Use:     "connection",
	Short:   "Connections: create, init, test, get, list, update, delete",
	Aliases: []string{"con"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(connectionCmd)
}

func init() {
	// Subcommands.
	connectionCmd.AddCommand(createCmd)
	connectionCmd.AddCommand(deleteCmd)
	connectionCmd.AddCommand(getCmd)
	connectionCmd.AddCommand(listCmd)
	connectionCmd.AddCommand(initCmd)
	connectionCmd.AddCommand(testCmd)
	// TODO: connectionCmd.AddCommand(updateCmd)
}

func connections() sdkservices.Connections {
	return common.Client().Connections()
}
