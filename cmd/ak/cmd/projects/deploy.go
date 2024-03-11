package projects

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var env string

var deployCmd = common.StandardCommand(&cobra.Command{
	Use:   "deploy <project name or ID> --from <file or directory> [--from ...] [--env <name or ID>]",
	Short: "Build, deploy, and activate project",
	Long:  `Build, deploy, and activate project - see also the "build" and "deployment" parent commands`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}

		// First, build the project (see the "build" sibling command).
		buildID, err := buildProject(args)
		if err != nil {
			return err
		}

		// Then, parse the optional environment argument.
		e, eid, err := r.EnvNameOrID(env, args[0])
		if err != nil {
			return err
		}
		if !e.IsValid() {
			err = fmt.Errorf("environment %q not found", env)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		// Finally, deploy and activate it (see the "deployment" parent command).
		deployment, err := sdktypes.DeploymentFromProto(&sdktypes.DeploymentPB{
			EnvId:   eid.String(),
			BuildId: buildID,
		})
		if err != nil {
			return fmt.Errorf("invalid deployment: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		did, err := deployments().Create(ctx, deployment)
		if err != nil {
			return fmt.Errorf("create deployment: %w", err)
		}

		common.RenderKV("deployment_id", did)

		if err := deployments().Activate(ctx, did); err != nil {
			return fmt.Errorf("activate deployment: %w", err)
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	deployCmd.Flags().StringArrayVarP(&paths, "from", "f", []string{}, "1 or more file or directory paths")
	kittehs.Must0(deployCmd.MarkFlagRequired("from"))

	deployCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
}

func deployments() sdkservices.Deployments {
	return common.Client().Deployments()
}
