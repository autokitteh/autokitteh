package envs

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--project=...] [--fail]",
	Short:   "List all execution environments",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			p   sdktypes.Project
			pid sdktypes.ProjectID
			err error
		)

		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		if project != "" {
			p, pid, err = r.ProjectNameOrID(ctx, project)
			if err = common.AddNotFoundErrIfCond(err, p.IsValid()); err != nil {
				return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "project")
			}
		}

		es, err := envs().List(ctx, pid)
		err = common.AddNotFoundErrIfCond(err, len(es) > 0)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "environments"); err == nil {
			common.RenderList(es)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)
}
