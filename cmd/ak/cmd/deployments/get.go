package deployments

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <deployment ID> [--fail]",
	Short:   "Get deployment details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		d, _, err := r.DeploymentID(args[0])
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "deployment", d); err != nil {
			return err
		}

		// Make the output deterministic during CLI integration tests.
		if test, err := cmd.Root().PersistentFlags().GetBool("test"); err == nil && test {
			d = sdktypes.DeploymentWithoutTimes(d)
		}

		common.RenderKVIfV("deployment", d)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
