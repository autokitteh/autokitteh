package triggers

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	call, event, filter, name, project, schedule string
	webhook                                      bool
)

var createCmd = common.StandardCommand(&cobra.Command{
	Use: `create -n name [--call file:func] [-p project] -c connection [-E event] [-f filter]
             create -n name [--call file:func] [-p project] -s "schedule"
             create -n name [--call file:func] [-p project] --webhook
`,

	Short: "Create event trigger",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		// Entry-point code location to call is required.
		cl, err := sdktypes.ParseCodeLocation(call)
		if err != nil {
			return fmt.Errorf("invalid entry-point to call %q: %w", call, err)
		}

		// Project is required.
		if project == "" {
			return errors.New("missing project")
		}

		pid, err := r.ProjectNameOrID(ctx, sdktypes.InvalidOrgID, project)
		if err = common.AddNotFoundErrIfCond(err, pid.IsValid()); err != nil {
			return common.WrapError(err, "project")
		}

		t, err := sdktypes.TriggerFromProto(&sdktypes.TriggerPB{
			Name:         name,
			EventType:    event,
			Filter:       filter,
			CodeLocation: cl.ToProto(),
			ProjectId:    pid.String(),
		})
		if err != nil {
			return fmt.Errorf("invalid trigger proto: %w", err)
		}

		if connection != "" {
			_, cid, err := r.ConnectionNameOrID(ctx, connection, project, sdktypes.InvalidOrgID)
			if err != nil {
				return common.WrapError(err, "connection")
			}

			t = t.WithConnectionID(cid)
		} else if schedule != "" {
			// TODO(ENG-1004): Verify validity of schedule expression.
			t = t.WithSchedule(schedule)
		} else if webhook {
			t = t.WithWebhook()
		} else {
			return errors.New("missing connection, schedule or webhook")
		}

		tid, err := triggers().Create(ctx, t)
		if err != nil {
			return fmt.Errorf("create trigger: %w", err)
		}

		common.RenderKVIfV("trigger_id", tid)
		return nil
	},
})

func init() {
	// Command-specific flags.
	createCmd.Flags().StringVarP(&name, "name", "n", "", "trigger name")
	kittehs.Must0(createCmd.MarkFlagRequired("name"))

	createCmd.Flags().StringVarP(&call, "call", "l", "", `entry-point to call ("filename:function")`)

	createCmd.Flags().VarP(common.NewNonEmptyString("", &project), "project", "p", "project name or ID")

	createCmd.Flags().VarP(common.NewNonEmptyString("", &connection), "connection", "c", "connection name or ID")
	createCmd.Flags().BoolVarP(&webhook, "webhook", "w", false, "trigger uses a webhook")

	createCmd.Flags().VarP(common.NewNonEmptyString("", &schedule), "schedule", "s", "schedule expression (cron or extended)")
	createCmd.MarkFlagsOneRequired("connection", "schedule", "webhook")
	createCmd.MarkFlagsMutuallyExclusive("connection", "schedule", "webhook")

	createCmd.Flags().StringVarP(&event, "event", "E", "", "optional event type, based on connection")
	createCmd.Flags().StringVarP(&filter, "filter", "f", "", "optional event data filter expression")
	createCmd.MarkFlagsMutuallyExclusive("schedule", "webhook", "event")
	createCmd.MarkFlagsMutuallyExclusive("schedule", "webhook")
	createCmd.MarkFlagsOneRequired("event", "schedule", "webhook")
}
