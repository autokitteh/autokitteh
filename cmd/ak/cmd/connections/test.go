package connections

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var testCmd = common.StandardCommand(&cobra.Command{
	Use:     "test <connection name or ID>",
	Short:   "Test connection",
	Aliases: []string{"t"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		_, cid, err := r.ConnectionNameOrID(ctx, args[0], "", sdktypes.InvalidOrgID)
		if err != nil {
			return err
		}

		s, err := connections().Test(ctx, cid)
		if err != nil {
			return fmt.Errorf("test connection: %w", err)
		}

		common.RenderKV("status", s)

		return nil
	},
})
