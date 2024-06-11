package triggers

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	call, event, filter, name, project, schedule string

	data map[string]string
)

var createCmd = common.StandardCommand(&cobra.Command{
	Use: `create -n name --call file:func [-p project] [-e env] -c connection [-E event] [-f filter] [--data key=value]...
             create -n name --call file:func [-p project] [-e env] -s "schedule"`,

	Short: "Create event trigger",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}

		// Entry-point code location to call is required.
		cl, err := sdktypes.ParseCodeLocation(call)
		if err != nil {
			return fmt.Errorf("invalid entry-point to call %q: %w", call, err)
		}

		// Project and/or environment are required.
		p, _, err := r.ProjectNameOrID(project)
		if err != nil {
			return err
		}
		if project != "" && !p.IsValid() {
			err = resolver.NotFoundError{Type: "project", Name: project}
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		e, eid, err := r.EnvNameOrID(env, project)
		if err != nil {
			return err
		}
		if env != "" && !e.IsValid() {
			err = resolver.NotFoundError{Type: "environment", Name: env}
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		// Mode 1: connection is required, event and/or filter
		// and/or data are required, and schedule is not allowed.
		c, cid, err := r.ConnectionNameOrID(connection, project)
		if connection != "" {
			if err != nil {
				if errors.As(err, resolver.NotFoundErrorType) {
					err = common.NewExitCodeError(common.NotFoundExitCode, err)
				}
				return err
			}
			if !c.IsValid() {
				err = resolver.NotFoundError{Type: "connection", Name: connection}
				return common.NewExitCodeError(common.NotFoundExitCode, err)
			}
		}

		// Mode 2: schedule is required, connection/event/filter are not allowed.
		// TODO(ENG-1004): Verify validity of schedule expression.
		if schedule != "" {
			cid = sdktypes.BuiltinSchedulerConnectionID
			event = fixtures.SchedulerEventTriggerType
			data[fixtures.ScheduleExpression] = schedule
			connection = fixtures.SchedulerConnectionName
		}

		// Finally, create and print the trigger.
		m := kittehs.TransformMapValues(data, func(v string) *sdktypes.ValuePB {
			return sdktypes.ToProto(sdktypes.NewStringValue(v))
		})

		t, err := sdktypes.StrictTriggerFromProto(&sdktypes.TriggerPB{
			Name:         name,
			EnvId:        eid.String(),
			ConnectionId: cid.String(),
			EventType:    event,
			Filter:       filter,
			Data:         m,
			CodeLocation: cl.ToProto(),
		})
		if err != nil {
			return fmt.Errorf("invalid trigger: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

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
	kittehs.Must0(createCmd.MarkFlagRequired("call"))

	createCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	createCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	createCmd.MarkFlagsOneRequired("project", "env")

	createCmd.Flags().StringVarP(&connection, "connection", "c", "", "connection name or ID")
	createCmd.Flags().StringVarP(&schedule, "schedule", "s", "", "schedule expression (cron or extended)")
	createCmd.MarkFlagsOneRequired("connection", "schedule")
	createCmd.MarkFlagsMutuallyExclusive("connection", "schedule")

	createCmd.Flags().StringVarP(&event, "event", "E", "", "optional event type, based on connection")
	createCmd.Flags().StringVarP(&filter, "filter", "f", "", "optional event data filter expression")
	createCmd.Flags().StringToStringVarP(&data, "data", "d", map[string]string{}, "optional event config key-value pairs")
	createCmd.MarkFlagsMutuallyExclusive("schedule", "event")
	createCmd.MarkFlagsMutuallyExclusive("schedule", "filter")
	createCmd.MarkFlagsMutuallyExclusive("schedule", "data")
	createCmd.MarkFlagsOneRequired("event", "filter", "data", "schedule")
}
