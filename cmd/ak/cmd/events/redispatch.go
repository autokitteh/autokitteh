package events

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var redispatchCmd = common.StandardCommand(&cobra.Command{
	Use:     "redispatch <event ID>",
	Short:   "Notify server's dispatcher about existing event",
	Aliases: []string{"red"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		e, _, err := r.EventID(args[0])
		if err != nil {
			return err
		}
		if e == nil {
			err = fmt.Errorf("event ID %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		eid, err := common.Client().Dispatcher().Dispatch(ctx, e, nil)
		if err != nil {
			return err
		}

		common.RenderKVIfV("event_id", eid)
		return nil
	},
})
