package builds

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <build ID>",
	Short:   "Delete build from server",
	Aliases: []string{"del", "de"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		b, id, err := r.BuildID(args[0])
		if err != nil {
			return err
		}
		if !b.IsValid() {
			err = fmt.Errorf("build ID %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		if err := builds().Delete(ctx, id); err != nil {
			return fmt.Errorf("delete build: %w", err)
		}

		return nil
	},
})
