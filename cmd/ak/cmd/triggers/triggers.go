package triggers

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flags shared by the "create" and "list" subcommands.
var env, connection string

var triggersCmd = common.StandardCommand(&cobra.Command{
	Use:     "triggers",
	Short:   "Event trigger management commands",
	Aliases: []string{"trigger", "trig"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(triggersCmd)
}

func init() {
	// Subcommands.
	triggersCmd.AddCommand(createCmd)
	triggersCmd.AddCommand(deleteCmd)
	triggersCmd.AddCommand(getCmd)
	triggersCmd.AddCommand(listCmd)
}

func triggers() sdkservices.Triggers {
	return common.Client().Triggers()
}
