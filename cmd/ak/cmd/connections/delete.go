package connections

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <connection name or ID>",
	Short:   "Discard existing connection to integration",
	Aliases: []string{"del", "d"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		_, id, err := r.ConnectionNameOrID(ctx, args[0], "", sdktypes.InvalidOrgID)
		if err != nil {
			return common.WrapError(err, "connection")
		}
		if !id.IsValid() {
			return common.WrapError(sdkerrors.ErrNotFound, "connection")
		}

		err = connections().Delete(ctx, id)
		return common.WrapError(err, "delete connection")
	},
})
