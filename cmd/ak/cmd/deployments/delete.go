package deployments

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <deployment ID>",
	Short:   "Delete inactive deployment",
	Aliases: []string{"del"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		_, did, err := r.DeploymentID(ctx, args[0])
		if err != nil {
			return common.WrapError(err)
		}

		if !did.IsValid() {
			return common.WrapError(sdkerrors.ErrNotFound, "deployment")
		}

		return common.WrapError(deployments().Delete(ctx, did))
	},
})
