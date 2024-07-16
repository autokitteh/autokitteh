package records

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list <event ID> [--fail]",
	Short:   "List all event records",
	Aliases: []string{"ls", "l"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		e, id, err := r.EventID(ctx, args[0])
		if err = common.AddNotFoundErrIfCond(err, e.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "event")
		}

		f := sdkservices.ListEventRecordsFilter{EventID: id}
		ers, err := events().ListEventRecords(ctx, f)
		err = common.AddNotFoundErrIfCond(err, len(ers) > 0)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "event records"); err == nil {
			common.RenderList(ers)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)
}
