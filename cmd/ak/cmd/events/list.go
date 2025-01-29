package events

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var listOrder, project string

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [filter flags] [--fail]",
	Short:   "List all events",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		var f sdkservices.ListEventsFilter

		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		if connection != "" {
			_, cid, err := r.ConnectionNameOrID(ctx, connection, "", sdktypes.InvalidOrgID)
			if err = common.AddNotFoundErrIfCond(err, cid.IsValid()); err != nil {
				return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "connection")
			}
			f.DestinationID = sdktypes.NewEventDestinationID(cid)
		}

		if trigger != "" {
			_, tid, err := r.TriggerNameOrID(ctx, sdktypes.InvalidOrgID, trigger, "")
			if err = common.AddNotFoundErrIfCond(err, tid.IsValid()); err != nil {
				return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "trigger")
			}
			f.DestinationID = sdktypes.NewEventDestinationID(tid)
		}

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

		f.Order = sdkservices.ListOrder(listOrder)

		es, err := events().List(ctx, f)
		err = common.AddNotFoundErrIfCond(err, len(es) > 0)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "events"); err == nil {
			common.RenderList(es)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&integration, "integration", "i", "", "integration name or ID")
	listCmd.Flags().StringVarP(&connection, "connection", "c", "", "connection name or ID")
	listCmd.Flags().StringVarP(&trigger, "trigger", "t", "", "trigger name or ID")
	listCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	listCmd.Flags().StringVarP(&eventType, "event-type", "e", "", "event type")
	listCmd.Flags().StringVarP(&listOrder, "order", "o", "DESC", "events order, should be DESC or ASC, default is DESC")

	common.AddFailIfNotFoundFlag(listCmd)
}
