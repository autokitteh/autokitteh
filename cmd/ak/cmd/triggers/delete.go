package triggers

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <trigger ID>",
	Short:   "Delete event trigger",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		t, id, err := r.TriggerID(args[0])
		if err != nil {
			return err
		}
		if !t.IsValid() {
			err = resolver.NotFoundError{Type: "trigger ID", Name: args[0]}
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		if err = triggers().Delete(ctx, id); err != nil {
			return fmt.Errorf("delete trigger: %w", err)
		}

		return nil
	},
})
