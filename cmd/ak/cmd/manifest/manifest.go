package manifest

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var quiet, rmUnusedConnVars bool

var manifestCmd = common.StandardCommand(&cobra.Command{
	Use:     "manifest",
	Short:   "Project manifest: schema, validate, plan, execute, apply, deploy",
	Aliases: []string{"man"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(manifestCmd)
}

func init() {
	// Subcommands.
	manifestCmd.AddCommand(applyCmd)
	manifestCmd.AddCommand(deployCmd)
	manifestCmd.AddCommand(execCmd)
	manifestCmd.AddCommand(planCmd)
	manifestCmd.AddCommand(schemaCmd)
	manifestCmd.AddCommand(validateCmd)
}

func logFunc(cmd *cobra.Command, prefix string) func(string) {
	if quiet {
		return func(string) {}
	}

	return func(msg string) {
		fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", prefix, msg)
	}
}
