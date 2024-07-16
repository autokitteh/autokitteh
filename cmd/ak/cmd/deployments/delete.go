package deployments

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <deployment ID> [--fail]",
	Short:   "Delete inactive deployment",
	Aliases: []string{"del"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		d, did, err := r.DeploymentID(ctx, args[0])
		if err = common.AddNotFoundErrIfCond(err, d.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "deployment")
		}

		return common.ToExitCodeWithSkipNotFoundFlag(cmd, deployments().Delete(ctx, did), "delete deployment")
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfError(deleteCmd)
}
