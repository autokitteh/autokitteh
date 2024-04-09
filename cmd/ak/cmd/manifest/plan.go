package manifest

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/manifest"
)

var noValidate, fromScratch bool

var planCmd = common.StandardCommand(&cobra.Command{
	Use:     "plan [file] [--no-validate] [--from-scratch] [--quiet]",
	Short:   "Dry-run for applying a YAML manifest, from a file or stdin",
	Aliases: []string{"p"},
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

		common.Render(actions)

		return nil
	},
})

func init() {
	planCmd.Flags().BoolVarP(&noValidate, "no-validate", "n", false, "do not validate")
	planCmd.Flags().BoolVarP(&fromScratch, "from-scratch", "s", false, "assume no existing setup")
	planCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "only show errors, if any")
}

func plan(cmd *cobra.Command, data []byte, path string) (manifest.Actions, error) {
	m, err := manifest.Read(data, path)
	if err != nil {
		return nil, err
	}

	ctx, cancel := common.LimitedContext()
	defer cancel()

	return manifest.Plan(
		ctx, m, common.Client(),
		manifest.WithLogger(logFunc(cmd, "plan")),
		manifest.WithFromScratch(fromScratch),
	)
}
