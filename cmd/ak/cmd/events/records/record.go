package records

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var recordCmd = common.StandardCommand(&cobra.Command{
	Use:   "record",
	Short: "Event record subcommands: add, list",
	Args:  cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(recordCmd)
}

func init() {
	// Subcommands.
	recordCmd.AddCommand(addCmd)
	recordCmd.AddCommand(listCmd)
}

func events() sdkservices.Events {
	return common.Client().Events()
}
