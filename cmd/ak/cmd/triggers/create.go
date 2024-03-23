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

var event, filter, loc string

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create <--env=...> <--connection=...> <--event=...> <--loc=...>",
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

		c, cid, err := r.ConnectionNameOrID(connection)
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

		cl, err := sdktypes.ParseCodeLocation(loc)
		if err != nil {
			return fmt.Errorf("invalid entry-point location %q: %w", loc, err)
		}

		t, err := sdktypes.StrictTriggerFromProto(&sdktypes.TriggerPB{
			EnvId:        eid.String(),
			ConnectionId: cid.String(),
			EventType:    event,
			Filter:       filter,
			CodeLocation: cl.ToProto(),
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

	createCmd.Flags().StringVarP(&connection, "connection", "n", "", "connection name or ID")
	kittehs.Must0(createCmd.MarkFlagRequired("connection"))

	createCmd.Flags().StringVarP(&event, "event", "E", "", "event type")
	kittehs.Must0(createCmd.MarkFlagRequired("event"))
	createCmd.Flags().StringVarP(&filter, "filter", "f", "", "event filter")

	createCmd.Flags().StringVarP(&loc, "loc", "l", "", "entrypoint code location")
	kittehs.Must0(createCmd.MarkFlagRequired("loc"))
}
