package builds

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <build ID> [--fail]",
	Short:   "Get build details from server",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		b, _, err := r.BuildID(ctx, args[0])
		err = common.AddNotFoundErrIfCond(err, b.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "build"); err == nil {
			common.RenderKVIfV("build", b)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
