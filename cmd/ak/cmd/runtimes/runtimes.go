package runtimes

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/backend/runtimes"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var local bool

func client() sdkservices.Runtimes {
	if local {
		return runtimes.New()
	}
	return common.Client().Runtimes()
}

var runtimesCmd = common.StandardCommand(&cobra.Command{
	Use:     "runtimes",
	Short:   "Runtime engine management commands",
	Aliases: []string{"runtime", "run", "rt", "r"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(runtimesCmd)
}

func init() {
	// Subcommands.
	runtimesCmd.AddCommand(getCmd)
	runtimesCmd.AddCommand(listCmd)
	runtimesCmd.AddCommand(buildCmd)

	runtimesCmd.PersistentFlags().BoolVarP(&local, "local", "l", false, "execute locally")
}
