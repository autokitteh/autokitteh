package triggers

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flags shared by the "create" and "list" subcommands.
var connection string

var triggerCmd = common.StandardCommand(&cobra.Command{
	Use:     "trigger",
	Short:   "Event triggers: create, get, list, delete",
	Aliases: []string{"trg"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(triggerCmd)
}

func init() {
	// Subcommands.
	triggerCmd.AddCommand(createCmd)
	triggerCmd.AddCommand(deleteCmd)
	triggerCmd.AddCommand(getCmd)
	triggerCmd.AddCommand(listCmd)
}

func triggers() sdkservices.Triggers {
	return common.Client().Triggers()
}
