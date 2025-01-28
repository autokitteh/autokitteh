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
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		f := sdkservices.ListDeploymentsFilter{}

		bid, err := sdktypes.ParseBuildID(buildID)
		if err != nil {
			return fmt.Errorf("invalid build ID %q: %w", buildID, err)
		}
		f.BuildID = bid

		if project != "" {
			pid, err := r.ProjectNameOrID(ctx, sdktypes.InvalidOrgID, project)
			if err = common.AddNotFoundErrIfCond(err, pid.IsValid()); err != nil {
				return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "project")
			}
			f.ProjectID = pid
		}

		if f.State, err = sdktypes.ParseDeploymentState(state.String()); err != nil {
			return fmt.Errorf("invalid state %q: %w", state, err)
		}

		f.IncludeSessionStats = includeSessionStats

		ds, err := deployments().List(ctx, f)
		err = common.AddNotFoundErrIfCond(err, len(ds) > 0)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "builds"); err == nil {
			// Make the output deterministic during CLI integration tests.
			if test, err := cmd.Root().PersistentFlags().GetBool("test"); err == nil && test {
				ds = kittehs.Transform(ds, func(d sdktypes.Deployment) sdktypes.Deployment { return d.WithoutTimestamps() })
			}
			common.RenderList(ds)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	listCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "build ID")
	listCmd.Flags().VarP(&state, "state", "s", strings.Join(possibleStates, "|"))
	listCmd.Flags().BoolVarP(&includeSessionStats, "include-session-stats", "i", false, "include session stats")

	common.AddFailIfNotFoundFlag(listCmd)
}
