package sessions

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get [session ID | project] [--fail]",
	Short:   "Get session configuration details (entry-point, inputs, etc.)",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		sid, err := acquireSessionID(args[0])
		if err = common.AddNotFoundErrIfCond(err, sid.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "session")
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		s, err := sessions().Get(ctx, sid)
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
