package connections

import (
	"errors"

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
		c, _, err := r.ConnectionNameOrID(args[0])
		if err != nil {
			if errors.As(err, resolver.NotFoundErrorType) {
				if err := common.FailIfNotFound(cmd, "connection", c); err != nil {
					return err
				}
				return nil
			}
			return err
		}

		if err := common.FailIfNotFound(cmd, "connection", c); err != nil {
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
