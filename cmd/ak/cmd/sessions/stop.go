package sessions

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var (
	reason string
	force  bool
)

var stopCmd = common.StandardCommand(&cobra.Command{
	Use:   "stop [session ID] [--reason <...>] [--force]",
	Short: "Stop running session",
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
		if !s.IsValid() {
			err = fmt.Errorf("session ID %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		if err = sessions().Stop(ctx, id, reason, force); err != nil {
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
