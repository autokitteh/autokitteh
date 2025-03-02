package sessions

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	endState, endPrintRE string
	waitCreated          bool
)

var watchCmd = common.StandardCommand(&cobra.Command{
	Use:   "watch [sessions ID | project] [--fail] [--end-state <state>] [--end-print-re <re>] [--timeout <duration>] [--poll-interval <duration>] [--no-timestamps] [--quiet] [--prints-only] [--wait-created]",
	Short: "Watch session runtime logs (prints, calls, errors, state changes)",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		sid, err := acquireSessionID(args[0])
		if err = common.AddNotFoundErrIfCond(err, sid.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "session")
		}

		endState, err := sdktypes.ParseSessionStateType(endState)
		if err != nil {
			return fmt.Errorf("end state: %w", err)
		}

		_, err = sessionWatch(sid, endState, endPrintRE)
		return err
	},
})

func init() {
	// Command-specific flags.
	watchCmd.Flags().StringVarP(&endState, "end-state", "e", "", "stop watching when state is reached")
	watchCmd.Flags().StringVarP(&endPrintRE, "end-print-re", "r", "", "stop watching when a print matching regex is reached")

	watchCmd.Flags().DurationVarP(&watchTimeout, "timeout", "t", 0, "timeout duration")
	watchCmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", defaultPollInterval, "poll interval")
	watchCmd.Flags().BoolVarP(&noTimestamps, "no-timestamps", "n", false, "omit timestamps from output")
	watchCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "don't print anything, just wait to finish")
	watchCmd.Flags().BoolVarP(&printsOnly, "just-prints", "p", false, "output only session print messages")

	watchCmd.Flags().BoolVar(&waitCreated, "wait-created", false, "wait for session to exist")

	common.AddFailIfNotFoundFlag(watchCmd)
}

func sessionWatch(sid sdktypes.SessionID, endState sdktypes.SessionStateType, endPrintRE string) ([]sdktypes.SessionLogRecord, error) {
	matchPrint := func(string) bool { return false }

	if endPrintRE != "" {
		pre, err := regexp.Compile(endPrintRE)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
		matchPrint = pre.MatchString
	}

	var state sdktypes.SessionStateType
	var rs []sdktypes.SessionLogRecord

	f := sdkservices.SessionLogRecordsFilter{SessionID: sid}
	f.PageSize = int32(pageSize)
	f.Ascending = true

	ctx := context.Background()
	if watchTimeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), watchTimeout)
		defer cancel()
	}

	first := true

	for !state.IsFinal() && (endState.IsZero() || state != endState) {
		if !first {
			time.Sleep(pollInterval)
		}

		first = false

		currCtx, cancel := common.WithLimitedContext(ctx)
		defer cancel()

		s, err := sessions().Get(currCtx, sid)
		if err != nil {
			cancel()

			if waitCreated && errors.Is(err, sdkerrors.ErrNotFound) {
				// session might not have been created yet (in test mode we know
				// in advance the session id).
				continue
			}

			return nil, fmt.Errorf("get session: %w", err)
		}

		state = s.State()

		f.Skip = int32(len(rs))
		f.PageToken = ""
		first := true

		for first || f.PageToken != "" {
			first = false

			res, err := sessions().GetLog(currCtx, f)
			if err != nil {
				cancel()
				return nil, err
			}

			logs := res.Records

			for _, log := range logs {
				if p, ok := log.GetPrint(); ok {
					s, _ := p.ToString()
					if matchPrint(s) {
						cancel()
						return rs, nil
					}
				}
			}

			printLogs(logs)

			f.PageToken = res.NextPageToken

			rs = append(rs, logs...)
			f.Skip = 0
		}

		cancel()
	}

	return rs, nil
}
