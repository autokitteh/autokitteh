package connections

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <connection name or ID>",
	Short:   "Discard existing connection to integration",
	Aliases: []string{"del", "d"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		c, id, err := r.ConnectionNameOrID(args[0], "")
		if err != nil {
			return common.ToExitCodeError(err, "connection")
		}
		if !c.IsValid() {
			return common.ToExitCodeError(sdkerrors.ErrNotFound, "connection")
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		err = connections().Delete(ctx, id)
		return common.ToExitCodeError(err, "delete connection")
	},
})
