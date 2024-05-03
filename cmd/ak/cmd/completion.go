package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var completionCmd = common.StandardCommand(&cobra.Command{
	Use:       "completion {bash|zsh|fish|powershell}",
	Short:     "Generate shell autocomplete script",
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},

	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return RootCmd.GenBashCompletionV2(cmd.OutOrStdout(), true)
		case "zsh":
			return RootCmd.GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			return RootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			return RootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		default:
			return fmt.Errorf("unsupported/unrecognized shell: %s", args[0])
		}
	},
})
