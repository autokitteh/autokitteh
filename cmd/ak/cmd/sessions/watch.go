package sessions

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var endState string

var watchCmd = common.StandardCommand(&cobra.Command{
	Use:   "watch [sessions ID] [--fail] [--end-state <state>] [--timeout <duration>] [--poll-interval <duration>] [--no-timestamps] [--quiet] [--prints-only]",
	Short: "Watch session runtime logs (prints, calls, errors, state changes)",
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
		ctx, cancel := common.LimitedContext()
		defer cancel()

		s, sid, err := r.SessionID(ctx, args[0])
		if err = common.AddNotFoundErrIfCond(err, s.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "session")
		}

		endState, err := sdktypes.ParseSessionStateType(endState)
		if err != nil {
			return fmt.Errorf("end state: %w", err)
		}

		_, err = sessionWatch(sid, endState)
		return err
	},
})

func init() {
	// Command-specific flags.
	watchCmd.Flags().StringVarP(&endState, "end-state", "e", "", "stop watching when state is reached")

	watchCmd.Flags().DurationVarP(&watchTimeout, "timeout", "t", 0, "timeout duration")
	watchCmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", defaultPollInterval, "poll interval")
	watchCmd.Flags().BoolVarP(&noTimestamps, "no-timestamps", "n", false, "omit timestamps from output")
	watchCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "don't print anything, just wait to finish")
	watchCmd.Flags().BoolVarP(&printsOnly, "just-prints", "p", false, "output only session print messages")

	common.AddFailIfNotFoundFlag(watchCmd)
}

func sessionWatch(sid sdktypes.SessionID, endState sdktypes.SessionStateType) ([]sdktypes.SessionLogRecord, error) {
	var state sdktypes.SessionStateType
	var rs []sdktypes.SessionLogRecord

	f := sdkservices.ListSessionLogRecordsFilter{SessionID: sid}
	f.PageSize = int32(pageSize)
	f.Ascending = true

	ctx := context.Background()
	if watchTimeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), watchTimeout)
		defer cancel()
	}

	for !state.IsFinal() && (endState.IsZero() || state != endState) {
		if len(rs) > 0 {
			time.Sleep(pollInterval)
		}

		currCtx, cancel := common.WithLimitedContext(ctx)

		s, err := sessions().Get(currCtx, sid)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("get session: %w", err)
		}

		state = s.State()

		f.Skip = int32(len(rs))
		f.PageToken = ""
		res, err := sessions().GetLog(currCtx, f)
		if err != nil {
			cancel()
			return nil, err
		}

		logs := res.Log.Records()

		printLogs(logs)

		f.PageToken = res.NextPageToken

		rs = append(rs, logs...)
		f.Skip = 0

		for f.PageToken != "" {
			res, err = sessions().GetLog(currCtx, f)
			if err != nil {
				cancel()
				return nil, err
			}

			logs := res.Log.Records()
			printLogs(logs)

			rs = append(rs, logs...)

			f.PageToken = res.NextPageToken
		}

		cancel()
	}

	return rs, nil
}
