package triggers

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:   "get <trigger ID> [--fail]",
	Short: "Get event trigger details",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		t, _, err := r.TriggerID(ctx, args[0])
		err = common.AddNotFoundErrIfNeeded(err, t.IsValid())
		if err = common.FailIfError2(cmd, err, "trigger"); err != nil {
			return err
		}

		common.RenderKVIfV("trigger", t)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
