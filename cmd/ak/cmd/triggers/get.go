package triggers

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <trigger ID> [--fail]",
	Short:   "Get event trigger details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		t, _, err := r.TriggerID(args[0])
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "trigger", t.IsValid()); err != nil {
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
