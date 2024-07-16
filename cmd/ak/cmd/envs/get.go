package envs

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <environment name or ID> [--project=...] [--fail]",
	Short:   "Get execution environment details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		e, _, err := r.EnvNameOrID(ctx, args[0], project)
		err = common.AddNotFoundErrIfCond(err, e.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "env"); err == nil {
			common.RenderKVIfV("env", e)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
