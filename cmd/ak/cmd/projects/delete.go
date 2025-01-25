package projects

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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

		pid, err := r.ProjectNameOrID(ctx, sdktypes.InvalidOrgID, args[0])
		if err = common.AddNotFoundErrIfCond(err, pid.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "project")
		}

		return common.ToExitCodeWithSkipNotFoundFlag(cmd, projects().Delete(ctx, pid), "delete project")
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfError(deleteCmd)
}
