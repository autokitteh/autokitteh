package records

import (
	"fmt"

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
		if err != nil {
			return err
		}
		if !e.IsValid() {
			err = fmt.Errorf("event ID %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		f := sdkservices.ListEventRecordsFilter{EventID: id}

		ers, err := events().ListEventRecords(ctx, f)
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "event records", len(ers) > 0); err != nil {
			return err
		}

		common.RenderList(ers)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)
}
