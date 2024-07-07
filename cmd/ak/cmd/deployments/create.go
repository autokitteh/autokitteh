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
	Use:     "create <--build-id=...> <--env=...> [--activate]",
	Short:   "Create new deployment",
	Aliases: []string{"c"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		b, _, err := r.BuildID(ctx, buildID)
		if err != nil {
			return err
		}
		if !b.IsValid() {
			err = fmt.Errorf("build ID %q not found", buildID)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		e, eid, err := r.EnvNameOrID(ctx, env, "")
		if err != nil {
			return err
		}
		if !e.IsValid() {
			err = fmt.Errorf("environment %q not found", env)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		deployment, err := sdktypes.DeploymentFromProto(&sdktypes.DeploymentPB{
			EnvId:   eid.String(),
			BuildId: buildID,
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

	createCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	kittehs.Must0(createCmd.MarkFlagRequired("env"))

	createCmd.Flags().BoolVarP(&activate, "activate", "a", false, "auto-activate deployment")
}
