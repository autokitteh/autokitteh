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
		s, _, err := r.SessionID(args[0])
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "session", s); err != nil {
			return err
		}

		common.RenderKVIfV("session", s)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
