package sessions

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get [session ID] [--fail]",
	Short:   "Get session configuration details (entry-point, inputs, etc.)",
	Aliases: []string{"g"},
	Args:    cobra.MaximumNArgs(1),

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

		s, _, err := r.SessionID(ctx, args[0])
		err = common.AddNotFoundErrIfCond(err, s.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "session"); err == nil {
			common.RenderKVIfV("session", s)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
