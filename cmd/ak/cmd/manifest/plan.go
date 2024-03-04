package manifest

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/manifest"
)

var noValidate, fromScratch bool

var planCmd = common.StandardCommand(&cobra.Command{
	Use:     "plan [file] [--no-validate] [--from-scratch]",
	Short:   "Dry-run for applying a YAML manifest, from a file or stdin",
	Aliases: []string{"p"},
	Args:    cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		data, path, err := common.Consume(args)
		if err != nil {
			return err
		}

		actions, err := plan(data, path)
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
}

func plan(data []byte, path string) (manifest.Actions, error) {
	m, err := manifest.Read(data, path)
	if err != nil {
		return nil, err
	}

	ctx, cancel := common.LimitedContext()
	defer cancel()

	return manifest.Plan(
		ctx,
		m,
		common.Client(),
		manifest.WithLogger(func(msg string) { fmt.Fprintf(os.Stderr, "[plan] %s\n", msg) }),
		manifest.WithFromScratch(fromScratch),
	)
}
