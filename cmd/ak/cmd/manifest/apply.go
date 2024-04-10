package manifest

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/manifest"
)

var applyCmd = common.StandardCommand(&cobra.Command{
	Use:     "apply [file] [--no-validate] [--from-scratch] [--quiet]",
	Short:   "Apply project configuration from file or stdin",
	Aliases: []string{"a"},
	Args:    cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		data, path, err := common.Consume(args)
		if err != nil {
			return err
		}

		actions, err := plan(cmd, data, path)
		if err != nil {
			return err
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		if _, err := manifest.Execute(ctx, actions, common.Client(), logFunc(cmd, "exec")); err != nil {
			return err
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	applyCmd.Flags().BoolVarP(&noValidate, "no-validate", "n", false, "do not validate")
	applyCmd.Flags().BoolVarP(&fromScratch, "from-scratch", "s", false, "assume no existing setup")
	applyCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "only show errors, if any")
}
