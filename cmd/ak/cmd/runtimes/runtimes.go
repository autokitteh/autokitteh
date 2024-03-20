package runtimes

import (
	"github.com/spf13/cobra"

	backendRuntimes "go.autokitteh.dev/autokitteh/backend/runtimes"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var local bool

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
	runtimesCmd.AddCommand(buildCmd)
	runtimesCmd.AddCommand(getCmd)
	runtimesCmd.AddCommand(listCmd)
	runtimesCmd.AddCommand(runCmd)
	runtimesCmd.AddCommand(testCmd)

	runtimesCmd.PersistentFlags().BoolVarP(&local, "local", "l", false, "execute locally")
}

func runtimes() sdkservices.Runtimes {
	if local {
		return backendRuntimes.New()
	}
	return common.Client().Runtimes()
}
