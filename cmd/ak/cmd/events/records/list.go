package records

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list <event ID> [--fail]",
	Short:   "List all event records",
	Aliases: []string{"ls", "l"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		e, id, err := r.EventID(args[0])
		if err != nil {
			return err
		}
		if e == nil {
			err = fmt.Errorf("event ID %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		f := sdkservices.ListEventRecordsFilter{EventID: id}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		ers, err := events().ListEventRecords(ctx, f)
		if err != nil {
			return err
		}

		if len(ers) == 0 {
			var dummy *sdktypes.EventRecord
			return common.FailIfNotFound(cmd, "event records", dummy)
		}

		common.RenderList(ers)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)
}
