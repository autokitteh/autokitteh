package events

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <event ID> [--fail]",
	Short:   "Get event details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		e, _, err := r.EventID(args[0])
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "event", e); err != nil {
			return err
		}

		common.RenderKVIfV("event", e)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
