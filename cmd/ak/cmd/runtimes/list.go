package runtimes

import (
	"context"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--fail]",
	Short:   "List all registered runtime engines",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		rs, err := runtimes().List(context.Background())
		err = common.AddNotFoundErrIfCond(err, len(rs) > 0)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "runtimes"); err == nil {
			common.RenderList(rs)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)
}
