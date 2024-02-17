package mappings

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flags shared by the "create" and "list" subcommands.
var env, connection string

var mappingsCmd = common.StandardCommand(&cobra.Command{
	Use:     "mappings",
	Short:   "Connection mapping management commands",
	Aliases: []string{"mapping", "map"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(mappingsCmd)
}

func init() {
	// Subcommands.
	mappingsCmd.AddCommand(createCmd)
	mappingsCmd.AddCommand(deleteCmd)
	mappingsCmd.AddCommand(getCmd)
	mappingsCmd.AddCommand(listCmd)
}

func mappings() sdkservices.Mappings {
	return common.Client().Mappings()
}
