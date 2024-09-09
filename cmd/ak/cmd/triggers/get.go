package triggers

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:   "get <trigger name or ID> [--fail] [--project project]",
	Short: "Get event trigger details",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		t, _, err := r.TriggerNameOrID(ctx, args[0], project)
		err = common.AddNotFoundErrIfCond(err, t.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "trigger"); err == nil {
			common.RenderKVIfV("trigger", t)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)

	getCmd.Flags().VarP(common.NewNonEmptyString("", &project), "project", "p", "project name or ID")
}
