package projects

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <project ID> [--fail]",
	Short:   "Delete project",
	Aliases: []string{"d"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		p, id, err := r.ProjectNameOrID(ctx, args[0])
		if err != nil {
			return common.FailIfError(cmd, err, "project")
		}

		if err := common.FailIfNotFound(cmd, "project", p.IsValid()); err != nil {
			return err
		}

		err = projects().Delete(ctx, id)
		return common.ToExitCodeError(err, "delete project")
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfError(deleteCmd)
}
