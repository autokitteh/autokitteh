package mappings

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <mapping ID>",
	Short:   "Delete connection mapping",
	Aliases: []string{"del", "d"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		m, id, err := r.MappingID(args[0])
		if err != nil {
			return err
		}
		if m == nil {
			err = fmt.Errorf("mapping ID %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		if err = mappings().Delete(ctx, id); err != nil {
			return fmt.Errorf("delete mapping: %w", err)
		}

		return nil
	},
})
