package connections

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <connection name or ID>",
	Short:   "Discard existing connection to integration",
	Aliases: []string{"d"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		c, id, err := r.ConnectionNameOrID(args[0])
		if err != nil {
			if errors.As(err, resolver.NotFoundErrorType) {
				err = common.NewExitCodeError(common.NotFoundExitCode, err)
			}
			return err
		}
		if c == nil {
			err = fmt.Errorf("connection %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		if err = connections().Delete(ctx, id); err != nil {
			return fmt.Errorf("delete connection: %w", err)
		}

		return nil
	},
})
