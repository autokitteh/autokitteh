package sessions

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var skip int

var logCmd = common.StandardCommand(&cobra.Command{
	Use:   "log [sessions ID] [--fail] [--skip N] [--no-timestamps]",
	Short: "Get session runtime logs (prints, calls, errors, state changes)",
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

		ctx, cancel := common.LimitedContext()
		defer cancel()

		_, err = sessionLog(ctx, id, skip)

		return err
	},
})

func init() {
	// Command-specific flags.
	logCmd.Flags().IntVar(&skip, "skip", 0, "number of entries to skip")
	logCmd.Flags().BoolVar(&noTimestamps, "no-timestamps", false, "omit timestamps from track output")

	common.AddFailIfNotFoundFlag(logCmd)
}

// skip >= 0: skip first records
// skip < 0: skip all up to last |skip| records.
func sessionLog(ctx context.Context, sid sdktypes.SessionID, skip int) ([]sdktypes.SessionLogRecord, error) {
	l, err := sessions().GetLog(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("get log: %w", err)
	}

	rs := l.Records()

	var fresh []sdktypes.SessionLogRecord

	if skip < 0 {
		fresh = rs[len(rs)+skip:]
	} else if len(rs) > skip {
		fresh = rs[skip:]
	}

	for _, r := range fresh {
		if noTimestamps {
			r = r.WithoutTimestamp()
		}

		common.Render(r)
	}

	return rs, nil
}
