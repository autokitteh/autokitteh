package sessions

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <session ID> [--fail]",
	Short:   "Delete non-running session",
	Aliases: []string{"d"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		s, sid, err := r.SessionID(ctx, args[0])
		if err = common.AddNotFoundErrIfCond(err, s.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "session")
		}

		if err = sessions().Delete(ctx, sid); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "delete session")
		}

		common.RenderKVIfV("session", s) // print deleted session
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfError(deleteCmd)
}
