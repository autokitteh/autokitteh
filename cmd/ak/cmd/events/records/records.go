package records

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var recordsCmd = common.StandardCommand(&cobra.Command{
	Use:     "records",
	Short:   "Event record subcommands",
	Aliases: []string{"record", "rec"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(recordsCmd)
}

func init() {
	// Subcommands.
	recordsCmd.AddCommand(addCmd)
	recordsCmd.AddCommand(listCmd)
}

func events() sdkservices.Events {
	return common.Client().Events()
}
