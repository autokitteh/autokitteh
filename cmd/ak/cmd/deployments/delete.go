package deployments

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <deployment ID> [--fail]",
	Short:   "Delete inactive deployment",
	Aliases: []string{"d"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		d, id, err := r.DeploymentID(args[0])
		if err != nil {
			return common.FailIfError(cmd, err, "deployment")
		}

		if err := common.FailIfNotFound(cmd, "deployment", d); err != nil { // test d != nil
			return err
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()
		err = deployments().Delete(ctx, id)
		if err != nil { // report any other than "not found" error (which was handled above)
			return common.ToExitCodeError(err, "delete deployment")
		}
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfError(deleteCmd)
}
