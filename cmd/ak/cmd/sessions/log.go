package sessions

import (
	"context"
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	// skip       int
	printsOnly bool
	logOrder   string
)

var logCmd = common.StandardCommand(&cobra.Command{
	Use:   "log [sessions ID] [--fail] [--skip <N>] [--no-timestamps] [--prints-only]",
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
		ctx, cancel := common.LimitedContext()
		defer cancel()

		s, id, err := r.SessionID(ctx, args[0])
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "session", s.IsValid()); err != nil {
			return err
		}

		f := sdkservices.ListSessionLogRecordsFilter{SessionID: id}
		if nextPageToken != "" {
			f.PageToken = nextPageToken
		}

		if pageSize > 0 {
			f.PageSize = int32(pageSize)
		}

		if skipRows > 0 {
			f.Skip = int32(skipRows)
		}

		f.Ascending = true
		if logOrder == "desc" {
			f.Ascending = false
		}

		return sessionLog(ctx, f)
	},
})

func init() {
	// Command-specific flags.
	// logCmd.Flags().IntVarP(&skip, "skip", "s", 0, "number of entries to skip")
	logCmd.Flags().BoolVarP(&noTimestamps, "no-timestamps", "n", false, "omit timestamps from watch output")
	logCmd.Flags().BoolVarP(&printsOnly, "prints-only", "p", false, "output only session print messages")
	logCmd.Flags().StringVarP(&logOrder, "order", "o", "desc", "logs order can be asc or desc")
	logCmd.Flags().StringVar(&nextPageToken, "next-page-token", "", "provide the returned page token to get next")
	logCmd.Flags().IntVar(&pageSize, "page-size", 20, "page size")
	logCmd.Flags().IntVar(&skipRows, "skip-rows", 0, "skip rows")

	common.AddFailIfNotFoundFlag(logCmd)
}

// skip >= 0: skip first records
// skip < 0: skip all up to last |skip| records.
func sessionLog(ctx context.Context, filter sdkservices.ListSessionLogRecordsFilter) error {
	l, err := sessions().GetLog(ctx, filter)
	if err != nil {
		return fmt.Errorf("get log: %w", err)
	}

	rs := l.Log.Records()
	if len(rs) == 0 {
		return nil
	}

	slices.SortFunc(rs, func(a, b sdktypes.SessionLogRecord) int {
		return a.Timestamp().Compare(b.Timestamp())
	})

	printLogs(rs)

	return nil
}

func printLogs(logs []sdktypes.SessionLogRecord) {
	for _, r := range logs {
		if noTimestamps {
			r = r.WithoutTimestamp().WithProcessID("")
		}

		msg := ""
		if printsOnly {
			if txt, ok := r.GetPrint(); ok {
				msg = txt
			} else if state := r.GetState(); state.IsValid() && state.Type() == sdktypes.SessionStateTypeError {
				if stateErr := state.GetError(); stateErr.IsValid() {
					pe := stateErr.GetProgramError()
					msg = fmt.Sprintf("Error: %s", pe.ErrorString())
				}
			}

			if msg != "" {
				if !noTimestamps {
					fmt.Printf("[%s] ", r.Timestamp().String())
				}
				fmt.Println(msg)
			}
			continue
		}

		if !quiet {
			common.Render(r)
		}
	}
}
