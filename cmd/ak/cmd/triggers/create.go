package triggers

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var event, filter, loc, name, schedule string

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     `create <--name=...> <--env=...> <--loc=...> { <--connection=...> <--event=...> | <--schedule=...> }`,
	Short:   "Create event trigger",
	Aliases: []string{"c"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		e, eid, err := r.EnvNameOrID(env, "")
		if err != nil {
			return err
		}
		if !e.IsValid() {
			err = fmt.Errorf("environment %q not found", env)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		cl, err := sdktypes.ParseCodeLocation(loc)
		if err != nil {
			return fmt.Errorf("invalid entry-point location %q: %w", loc, err)
		}

		connectionID := ""
		data := make(map[string]sdktypes.Value)
		if cmd.Flags().Changed("schedule") {
			if connection != "" || event != "" {
				return fmt.Errorf(`flags(s) "connection", "event" are not compatible with "schedule"`)
			}
			event = sdktypes.SchedulerEventTriggerType
			data[sdktypes.ScheduleExpression] = sdktypes.NewStringValue(schedule)
			connection = sdktypes.SchedulerConnectionName
			connectionID = sdktypes.BuiltinSchedulerConnectionID.String()
		} else {
			if connection == "" || event == "" {
				return fmt.Errorf(`required flag(s) "connection", "event" not set`)
			}
			c, cid, err := r.ConnectionNameOrID(connection, "")
			if err != nil {
				if errors.As(err, resolver.NotFoundErrorType) {
					err = common.NewExitCodeError(common.NotFoundExitCode, err)
				}
				return err
			}
			if !c.IsValid() {
				err = fmt.Errorf("connection %q not found", connection)
				return common.NewExitCodeError(common.NotFoundExitCode, err)
			}
			connectionID = cid.String()
		}

		t, err := sdktypes.StrictTriggerFromProto(&sdktypes.TriggerPB{
			EnvId:        eid.String(),
			ConnectionId: connectionID,
			EventType:    event,
			Filter:       filter,
			CodeLocation: cl.ToProto(),
			Name:         name,
			Data:         kittehs.TransformMapValues(data, sdktypes.ToProto),
		})
		if err != nil {
			return fmt.Errorf("invalid trigger: %w", err)
		}

		tid, err := triggers().Create(context.Background(), t)
		if err != nil {
			return fmt.Errorf("create trigger: %w", err)
		}

		common.RenderKVIfV("trigger_id", tid)
		return nil
	},
})

func init() {
	// Command-specific flags.
	createCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	kittehs.Must0(createCmd.MarkFlagRequired("env"))

	createCmd.Flags().StringVarP(&name, "name", "n", "", "trigger name")
	kittehs.Must0(createCmd.MarkFlagRequired("name"))

	createCmd.Flags().StringVarP(&connection, "connection", "c", "", "connection name or ID")

	createCmd.Flags().StringVarP(&event, "event", "E", "", "event type")

	createCmd.Flags().StringVarP(&filter, "filter", "f", "", "event filter")

	createCmd.Flags().StringVarP(&loc, "loc", "l", "", "entrypoint code location")
	kittehs.Must0(createCmd.MarkFlagRequired("loc"))

	createCmd.Flags().StringVarP(&schedule, "schedule", "s", "", "cron schedule")
}
