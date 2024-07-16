package integrations

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <integration name or ID> [--fail]",
	Short:   "Get integration details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		i, _, err := r.IntegrationNameOrID(ctx, args[0])
		err = common.AddNotFoundErrIfCond(err, i.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "integration"); err == nil {
			common.RenderKVIfV("integration", i)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
