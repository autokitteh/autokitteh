package sessions

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	track        bool
	pollInterval time.Duration
	noTimestamps bool
	endState     string
	watchTimeout time.Duration
	quiet        bool
)

var watchCmd = common.StandardCommand(&cobra.Command{
	Use:   "watch [sessions ID] [--fail] [--no-timestamps] [--poll-interval] [--end-state STATE] [--timeout DURATION] [--quiet]",
	Short: "Watch for session runtime logs (prints, calls, errors, state changes)",
	Args:  cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			id, err := latestSessionID()
			if err != nil {
				return err
			}
			args = append(args, id)
		}

		r := resolver.Resolver{Client: common.Client()}
		s, id, err := r.SessionID(args[0])
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "session", s.IsValid()); err != nil {
			return err
		}

		endState, err := sdktypes.ParseSessionStateType(endState)
		if err != nil {
			return fmt.Errorf("end state: %w", err)
		}

		_, err = sessionWatch(id, endState)
		return err
	},
})

func init() {
	// Command-specific flags.
	watchCmd.Flags().DurationVar(&pollInterval, "poll-interval", defaultPollInterval, "poll interval")
	watchCmd.Flags().BoolVar(&noTimestamps, "no-timestamps", false, "omit timestamps from track output")
	watchCmd.Flags().StringVar(&endState, "end-state", "", "stop watching when state is reached")
	watchCmd.Flags().DurationVar(&watchTimeout, "timeout", 0, "time out duration")
	watchCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "do not print anything, just wait to finish")

	common.AddFailIfNotFoundFlag(watchCmd)
}

func sessionWatch(sid sdktypes.SessionID, endState sdktypes.SessionStateType) ([]sdktypes.SessionLogRecord, error) {
	var state sdktypes.SessionStateType
	var rs []sdktypes.SessionLogRecord

	ctx := context.Background()
	if watchTimeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), watchTimeout)
		defer cancel()
	}

	for last := 0; !state.IsFinal() && (endState.IsZero() || state != endState); last = len(rs) {
		if last > 0 {
			time.Sleep(pollInterval)
		}

		currCtx, cancel := common.WithLimitedContext(ctx)

		s, err := sessions().Get(currCtx, sid)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("get session: %w", err)
		}

		state = s.State()

		if rs, err = sessionLog(currCtx, sid, last); err != nil {
			cancel()
			return nil, err
		}

		cancel()
	}

	return rs, nil
}
