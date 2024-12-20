package sessions

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var (
	reason string
	force  bool
)

var stopCmd = common.StandardCommand(&cobra.Command{
	Use:   "stop [session ID | project] [--reason <...>] [--force]",
	Short: "Stop running session",
	Args:  cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		sid, err := acquireSessionID(args[0])
		if err = common.AddNotFoundErrIfCond(err, sid.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "session")
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		if err = sessions().Stop(ctx, sid, reason, force); err != nil {
			return fmt.Errorf("stop session: %w", err)
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	stopCmd.Flags().StringVarP(&reason, "reason", "r", "", "optional reason for stopping")
	stopCmd.Flags().BoolVarP(&force, "force", "f", false, "terminate forcefully")
}
