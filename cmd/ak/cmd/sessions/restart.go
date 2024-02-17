package sessions

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var restartCmd = common.StandardCommand(&cobra.Command{
	Use:   "restart [session ID] [--wait]",
	Short: "Start new instance of existing session",
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
		s, _, err := r.SessionID(args[0])
		if err != nil {
			return err
		}
		if s == nil {
			err = fmt.Errorf("session ID %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		sid, err := sessions().Start(ctx, s)
		if err != nil {
			return fmt.Errorf("start session: %w", err)
		}

		common.RenderKVIfV("session_id", sid)

		if wait {
			fmt.Println()
			e := make(chan error)
			done := make(chan bool)
			state := make(chan string)
			go waitForSession(sid, e, done, state)
			go updateStateTicker(e, done, state)
			select {
			case err := <-e:
				return err
			case <-done:
				break
			}
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	restartCmd.Flags().BoolVarP(&wait, "wait", "w", false, "wait for session to complete")
	restartCmd.Flags().DurationVarP(&waitInterval, "wait-interval", "i", defaultWait, "wait interval")
}
