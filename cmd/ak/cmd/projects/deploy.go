package projects

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var deployCmd = common.StandardCommand(&cobra.Command{
	Use:   "deploy <project name or ID> [--dir <path> [...]] [--file <path> [...]]",
	Short: "Build, deploy, and activate project",
	Long:  `Build, deploy, and activate project - see also the "build" sibling and "deployment" parent commands`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		pid, err := r.ProjectNameOrID(ctx, args[0])
		if err != nil {
			err = fmt.Errorf("project: %w", err)

			if errors.Is(err, sdkerrors.ErrNotFound) {
				err = common.NewExitCodeError(common.NotFoundExitCode, err)
			}

			return err
		}

		// Step 1: build the project (see the "build" sibling command).
		bid, err := common.BuildProject(pid, dirPaths, filePaths)
		if err != nil {
			return err
		}
		common.RenderKV("build_id", bid)

		// Step 2: deploy the build (see the "deployment" parent command).
		deployment, err := sdktypes.DeploymentFromProto(&sdktypes.DeploymentPB{
			ProjectId: pid.String(),
			BuildId:   bid.String(),
		})
		if err != nil {
			return fmt.Errorf("invalid deployment: %w", err)
		}

		did, err := deployments().Create(ctx, deployment)
		if err != nil {
			return fmt.Errorf("create deployment: %w", err)
		}
		common.RenderKV("deployment_id", did)

		// Step 3: activate the deployment (see the "deployment" parent command).
		if err := deployments().Activate(ctx, did); err != nil {
			return fmt.Errorf("activate deployment: %w", err)
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	deployCmd.Flags().StringArrayVarP(&dirPaths, "dir", "d", []string{}, "0 or more directory paths")
	deployCmd.Flags().StringArrayVarP(&filePaths, "file", "f", []string{}, "0 or more file paths")
	kittehs.Must0(deployCmd.MarkFlagDirname("dir"))
	kittehs.Must0(deployCmd.MarkFlagFilename("file"))
	deployCmd.MarkFlagsOneRequired("dir", "file")
}

func deployments() sdkservices.Deployments {
	return common.Client().Deployments()
}
