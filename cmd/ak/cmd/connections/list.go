package connections

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--integration=...] [--project=...] [--fail]",
	Short:   "List all connections",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		var f sdkservices.ListConnectionsFilter

		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		if integration != "" {
			i, iid, err := r.IntegrationNameOrID(ctx, integration)
			if err = common.AddNotFoundErrIfCond(err, i.IsValid()); err != nil {
				return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "integration")
			}
			f.IntegrationID = iid
		}

		if project != "" {
			pid, err := r.ProjectNameOrID(ctx, sdktypes.InvalidOrgID, project)
			if err = common.AddNotFoundErrIfCond(err, pid.IsValid()); err != nil {
				return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "project")
			}
			f.ProjectID = pid
		}

		cs, err := connections().List(ctx, f)
		err = common.AddNotFoundErrIfCond(err, len(cs) > 0)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "connections"); err == nil {
			common.RenderList(cs)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&integration, "integration", "i", "", "integration name or ID")
	listCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")

	common.AddFailIfNotFoundFlag(listCmd)
}
