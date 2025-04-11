package sessions

import (
	"fmt"
	"regexp"
	"time"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var tail bool

var printsCmd = common.StandardCommand(&cobra.Command{
	Use:   "prints [sessions ID | project] [--fail] [--no-timestamps] [--poll-interval <duration>] [--tail] [--end-print-re <re>]",
	Short: "Get session prints",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		sid, err := acquireSessionID(args[0])
		if err = common.AddNotFoundErrIfCond(err, sid.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "session")
		}

		matchPrint := func(string) bool { return false }

		if endPrintRE != "" {
			pre, err := regexp.Compile(endPrintRE)
			if err != nil {
				return fmt.Errorf("invalid regex: %w", err)
			}
			matchPrint = pre.MatchString
		}

		var (
			n     int32
			first = true
			more  = false
		)

		for {
			if !first && !more {
				time.Sleep(pollInterval)
			}

			first = false

			ctx, done := common.LimitedContext()
			defer done()

			var s sdktypes.Session
			if tail {
				// We need the session state just in case of tail, to know if
				// there if the session ended and no further prints are coming.
				if s, err = sessions().Get(ctx, sid); err != nil {
					return fmt.Errorf("get session: %w", err)
				}
			}

			prints, err := sessions().GetPrints(ctx, sid, sdktypes.PaginationRequest{
				Ascending: true,
				Skip:      n,
			})
			if err != nil {
				return fmt.Errorf("get log: %w", err)
			}

			more = prints.NextPageToken != ""

			n += int32(len(prints.Prints))

			for _, p := range prints.Prints {
				text, err := p.Value.ToString()
				if err != nil {
					text = fmt.Sprintf("error converting print to string: %v", err.Error())
				}

				if !noTimestamps {
					fmt.Fprintf(cmd.OutOrStdout(), "[%s] ", p.Timestamp.String())
				}

				fmt.Fprintln(cmd.OutOrStdout(), text)

				if matchPrint(text) {
					return nil
				}
			}

			if !more && (!tail || s.State().IsFinal()) {
				break
			}
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	printsCmd.Flags().BoolVarP(&noTimestamps, "no-timestamps", "n", false, "omit timestamps from watch output")
	printsCmd.Flags().BoolVarP(&tail, "tail", "t", false, "follow the prints")
	printsCmd.Flags().StringVarP(&endPrintRE, "end-print-re", "r", "", "stop tail when a print matching regex is reached")
	printsCmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", defaultPollInterval, "poll interval")

	common.AddFailIfNotFoundFlag(printsCmd)
}
