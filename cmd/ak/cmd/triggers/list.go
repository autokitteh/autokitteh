package triggers

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [-p project] [-e env] [-c connection] [--fail]",
	Short:   "List all event triggers",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		f := sdkservices.ListTriggersFilter{}

		// All flags are optional.
		if project != "" {
			p, pid, err := r.ProjectNameOrID(ctx, project)
			if err = common.AddNotFoundErrIfCond(err, p.IsValid()); err != nil {
				return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "project")
			}
			f.ProjectID = pid
		}

		if connection != "" {
			c, cid, err := r.ConnectionNameOrID(ctx, connection, project)
			if err = common.AddNotFoundErrIfCond(err, c.IsValid()); err != nil {
				return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "connection")
			}
			f.ConnectionID = cid
		}

		ts, err := triggers().List(ctx, f)
		err = common.AddNotFoundErrIfCond(err, len(ts) > 0)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "triggers"); err == nil {
			common.RenderList(ts)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	listCmd.Flags().StringVarP(&connection, "connection", "c", "", "connection name or ID")

	common.AddFailIfNotFoundFlag(listCmd)
}
