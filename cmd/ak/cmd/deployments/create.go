package deployments

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var activate bool

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create <--build-id=...> <--project=...> [--activate]",
	Short:   "Create new deployment",
	Aliases: []string{"c"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		b, _, err := r.BuildID(ctx, buildID)
		err = common.AddNotFoundErrIfCond(err, b.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, fmt.Sprintf("build ID %q", buildID)); err != nil {
			return err
		}

		pid, err := r.ProjectNameOrID(ctx, sdktypes.InvalidOrgID, project)
		err = common.AddNotFoundErrIfCond(err, pid.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, fmt.Sprintf("project %q", project)); err != nil {
			return err
		}

		deployment, err := sdktypes.DeploymentFromProto(&sdktypes.DeploymentPB{
			ProjectId: pid.String(),
			BuildId:   buildID,
		})
		if err != nil {
			return fmt.Errorf("invalid deployment: %w", err)
		}

		did, err := deployments().Create(ctx, deployment)
		if err != nil {
			return fmt.Errorf("create deployment: %w", err)
		}

		common.RenderKV("deployment_id", did)

		if activate {
			if err := deployments().Activate(ctx, did); err != nil {
				return fmt.Errorf("activate deployment: %w", err)
			}
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	createCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "build ID")
	kittehs.Must0(createCmd.MarkFlagRequired("build-id"))

	createCmd.Flags().StringVarP(&project, "project", "e", "", "project name or ID")
	kittehs.Must0(createCmd.MarkFlagRequired("project"))

	createCmd.Flags().BoolVarP(&activate, "activate", "a", false, "auto-activate deployment")
}
