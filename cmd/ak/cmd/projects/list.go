package projects

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--fail]",
	Short:   "List all projects",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		ps, err := projects().List(ctx)
		err = common.AddNotFoundErrIfCond(err, len(ps) > 0)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "projects"); err == nil {
			common.RenderList(ps)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)
}
