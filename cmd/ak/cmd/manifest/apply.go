package manifest

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/manifest"
)

var dryRun bool

var applyCmd = common.StandardCommand(&cobra.Command{
	Use:     "apply [file] [--no-validate] [--from-scratch] [--dry-run]",
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

		if !dryRun {
			ctx, cancel := common.LimitedContext()
			defer cancel()

			_, err := manifest.Execute(ctx, actions, common.Client(), logFunc(cmd, "exec"))
			return err
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	applyCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "only show plan, don't apply")
	applyCmd.Flags().BoolVarP(&noValidate, "no-validate", "n", false, "do not validate")
	applyCmd.Flags().BoolVarP(&fromScratch, "from-scratch", "s", false, "assume no existing setup")
}
