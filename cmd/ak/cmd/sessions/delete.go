package sessions

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete session ID [--fail]",
	Short:   "Delete non-running session",
	Aliases: []string{"d"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		s, id, err := r.SessionID(args[0])
		if err != nil {
			return common.FailIfError(cmd, err, "session")
		}

		if err := common.FailIfNotFound(cmd, "session id", s); err != nil {
			return err
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()
		err = sessions().Delete(ctx, id)

		if err != nil {
			return common.FailIfError(cmd, err, "session")
		}

		common.RenderKVIfV("session", s) // print deleted session
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfError(deleteCmd)
}
