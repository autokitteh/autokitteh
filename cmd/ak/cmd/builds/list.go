package builds

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--fail]",
	Short:   "List all uploaded project builds",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		bs, err := builds().List(ctx, sdkservices.ListBuildsFilter{})
		err = common.AddNotFoundErrIfCond(err, len(bs) > 0)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "builds"); err == nil {
			common.RenderList(bs)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)
}
