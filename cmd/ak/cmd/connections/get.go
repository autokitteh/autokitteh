package connections

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <connection name or ID> [--fail]",
	Short:   "Get connection details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		c, _, err := r.ConnectionNameOrID(ctx, args[0], "", sdktypes.InvalidOrgID)
		err = common.AddNotFoundErrIfCond(err, c.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "connection"); err == nil {
			common.RenderKVIfV("connection", c)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
