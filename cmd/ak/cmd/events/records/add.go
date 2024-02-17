package records

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var state stateString

var addCmd = common.StandardCommand(&cobra.Command{
	Use:     "add <event ID> [state=...]",
	Short:   "Add event record",
	Aliases: []string{"a"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		e, _, err := r.EventID(args[0])
		if err != nil {
			return err
		}
		if e == nil {
			err = fmt.Errorf("event ID %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		record, err := sdktypes.EventRecordFromProto(&sdktypes.EventRecordPB{
			EventId: args[0],
			State:   sdktypes.ParseEventRecordState(state.String()).ToProto(),
		})
		if err != nil {
			return fmt.Errorf("invalid event record: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		err = events().AddEventRecord(ctx, record)
		if err != nil {
			return fmt.Errorf("add event records: %w", err)
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	addCmd.Flags().VarP(&state, "state", "s", strings.Join(possibleStates, "|"))
}
