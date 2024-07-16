package deployments

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <deployment ID> [--fail]",
	Short:   "Get deployment details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		d, _, err := r.DeploymentID(ctx, args[0])
		err = common.AddNotFoundErrIfCond(err, d.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "deployment"); err == nil {
			// Make the output deterministic during CLI integration tests.
			if test, err := cmd.Root().PersistentFlags().GetBool("test"); err == nil && test {
				d = d.WithoutTimestamps()
			}
			common.RenderKVIfV("deployment", d)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
