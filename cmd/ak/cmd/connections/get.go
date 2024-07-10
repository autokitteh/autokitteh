package connections

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
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

		c, _, err := r.ConnectionNameOrID(ctx, args[0], "")
		err = common.AddNotFoundErrIfCond(err, c.IsValid())
		if err = common.FailIfError2(cmd, err, "connection"); err != nil {
			return err
		}

		common.RenderKVIfV("connection", c)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
