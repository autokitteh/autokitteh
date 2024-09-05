package triggers

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <trigger name or ID> [--project project]",
	Short:   "Delete event trigger",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		t, id, err := r.TriggerNameOrID(ctx, args[0], project)
		if err != nil {
			return err
		}
		if !t.IsValid() {
			err = resolver.NotFoundError{Type: "trigger ID", Name: args[0]}
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		if err = triggers().Delete(ctx, id); err != nil {
			return fmt.Errorf("delete trigger: %w", err)
		}

		return nil
	},
})

func init() {
	deleteCmd.Flags().VarP(common.NewNonEmptyString("", &project), "project", "p", "project name or ID")
}
