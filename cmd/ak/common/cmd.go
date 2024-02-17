package common

import (
	"github.com/spf13/cobra"
)

func StandardCommand(cmd *cobra.Command) *cobra.Command {
	// Don't add "[flags]" suffix to usage line.
	cmd.DisableFlagsInUseLine = true
	// Deal with errors explicitly in main.
	cmd.SilenceErrors = true
	// No need to print usage on every error.
	cmd.SilenceUsage = true
	// Don't sort flags in help messages.
	cmd.Flags().SortFlags = false

	return cmd
}
