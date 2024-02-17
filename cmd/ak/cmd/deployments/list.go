package deployments

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	state               stateString
	includeSessionStats bool
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [filter flags] [--fail]",
	Short:   "List all deployments",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		f := sdkservices.ListDeploymentsFilter{}

		bid, err := sdktypes.ParseBuildID(buildID)
		if err != nil {
			return fmt.Errorf("invalid build ID %q: %w", buildID, err)
		}
		f.BuildID = bid

		if env != "" {
			r := resolver.Resolver{Client: common.Client()}
			e, _, err := r.EnvNameOrID(env, "")
			if err != nil {
				return err
			}
			f.EnvID = sdktypes.GetEnvID(e)
		}

		f.State = sdktypes.ParseDeploymentState(state.String())
		f.IncludeSessionStats = includeSessionStats

		ctx, cancel := common.LimitedContext()
		defer cancel()

		ds, err := deployments().List(ctx, f)
		if err != nil {
			return fmt.Errorf("list deployments: %w", err)
		}

		if len(ds) == 0 {
			var dummy *sdktypes.Deployment
			return common.FailIfNotFound(cmd, "deployments", dummy)
		}

		// Make the output deterministic during CLI integration tests.
		if test, err := cmd.Root().PersistentFlags().GetBool("test"); err == nil && test {
			ds = kittehs.Transform(ds, sdktypes.DeploymentWithoutTimes)
		}

		common.RenderList(ds)
		return nil
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	listCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "build ID")
	listCmd.Flags().VarP(&state, "state", "s", strings.Join(possibleStates, "|"))
	listCmd.Flags().BoolVarP(&includeSessionStats, "include-session-stats", "i", false, "include session stats")

	common.AddFailIfNotFoundFlag(listCmd)
}
