package sessions

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete session ID [--fail]",
	Short:   "Delete non-running session",
	Aliases: []string{"d"},
	Args:    cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		s, id, err := r.SessionID(args[0])
		if err != nil {
			return err
		}

		if s != nil {
			ctx, cancel := common.LimitedContext()
			defer cancel()

			common.RenderKVIfV("session", s) // FIXME: should we print deleted session?
			err = sessions().Delete(ctx, id)
			if err != nil {
				err = fmt.Errorf("delete session - id<%q>: %w", id, err)
			}
		} else {
			err = common.NewExitCodeError(common.NotFoundExitCode, fmt.Errorf("session id<%q> not found", id))
		}

		if kittehs.Must1(cmd.Flags().GetBool("fail")) && err != nil {
			return err
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(deleteCmd)
}
