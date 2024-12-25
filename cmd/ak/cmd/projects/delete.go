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

		p, pid, err := r.ProjectNameOrID(ctx, args[0])
		if err = common.AddNotFoundErrIfCond(err, p.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "project")
		}

		return common.ToExitCodeWithSkipNotFoundFlag(cmd, projects().Delete(ctx, pid), "delete project")
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfError(deleteCmd)
}
