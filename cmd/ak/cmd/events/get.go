package events

import (
	"errors"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <event ID> [--fail]",
	Short:   "Get event details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		e, _, err := r.EventID(ctx, args[0])
		if err != nil {
			if errors.Is(err, sdkerrors.ErrNotFound) {
				return common.FailNotFound(cmd, "event")
			}
			return err
		}

		if err := common.FailIfNotFound(cmd, "event", e.IsValid()); err != nil {
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
