package sessions

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var restartCmd = common.StandardCommand(&cobra.Command{
	Use:   "restart [session ID | project] [--watch [--watch-timeout <duration>] [--poll-interval <duration>] [--no-timestamps] [--quiet]]",
	Short: "Start new instance of existing session",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		sid, err := acquireSessionID(args[0])
		if err = common.AddNotFoundErrIfCond(err, sid.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "session")
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		s, err := sessions().Get(ctx, sid)
		if err != nil {
			return err
		}
		if !s.IsValid() {
			err = fmt.Errorf("session ID %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}
		s = s.WithNoID() // remove sessionID

		sid, err = sessions().Start(ctx, s)
		if err != nil {
			return fmt.Errorf("start session: %w", err)
		}

		common.RenderKVIfV("session_id", sid)

		if watch {
			_, err := sessionWatch(sid, sdktypes.SessionStateTypeUnspecified, "")
			return err
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	restartCmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch session to completion")

	restartCmd.Flags().DurationVarP(&watchTimeout, "watch-timeout", "t", 0, "watch timeout duration")
	restartCmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", defaultPollInterval, "watch poll interval")
	restartCmd.Flags().BoolVarP(&noTimestamps, "no-timestamps", "n", false, "omit timestamps from watch output")
	restartCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "don't print anything, just wait to finish")
}
