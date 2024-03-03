package sessions

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	filterInput  bool
	filterOutput bool
)

var logCmd = common.StandardCommand(&cobra.Command{
	Use:   "log [sessions ID] [--fail] [--filter-input] [--filter-output]",
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

		l, err := sessions().GetLog(ctx, id)
		if err != nil {
			return fmt.Errorf("session log: %w", err)
		}

		pb := l.ToProto()
		for _, r := range pb.Records {
			if filterInput {
				f := r.GetCallSpec().GetFunction().GetFunction()
				if f != nil {
					f.Data = nil
					f.Desc = nil
				}
			}
			if filterOutput {
				r.GetCallAttemptComplete().Result = nil
			}
		}

		if l, err = sdktypes.SessionLogFromProto(pb); err != nil {
			return fmt.Errorf("omit extra details: %w", err)
		}

		common.RenderKVIfV("log", l)
		return nil
	},
})

func init() {
	// Command-specific flags.
	logCmd.Flags().BoolVar(&filterInput, "filter-input", false, "filter input details")
	logCmd.Flags().BoolVar(&filterOutput, "filter-output", false, "filter output details")

	common.AddFailIfNotFoundFlag(logCmd)
}
