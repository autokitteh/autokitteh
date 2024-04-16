package manifest

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	env string

	filePaths, dirPaths []string
)

// Allow no dir/file - use manifest dir by default

var deployCmd = common.StandardCommand(&cobra.Command{
	Use:   "deploy <manifest file> [--project-name <name>] [--dir <path> [...]] [--file <path> [...]] [--env <name or ID>] [--quiet]",
	Short: "Create, configure, build, deploy, and activate project",
	Long:  `Create, configure, build, deploy, and activate project - see also the "build", "deployment", and "project" parent commands`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}

		// Step 1: apply the manifest file (see also the "manifest" parent command).
		project, err := applyManifest(cmd, args)
		if err != nil {
			return err
		}

		// Step 2: build the project
		// (see also the "build" and "project" parent commands).
		if len(dirPaths) == 0 && len(filePaths) == 0 {
			dirPaths = append(dirPaths, filepath.Dir(args[0]))
		}

		bid, err := common.BuildProject(project, dirPaths, filePaths)
		if err != nil {
			return err
		}
		logFunc(cmd, "exec")(fmt.Sprintf("create_build: created %q", bid))

		// Step 3: parse the optional environment argument.
		e, eid, err := r.EnvNameOrID(env, project)
		if err != nil {
			return err
		}
		if !e.IsValid() {
			err = fmt.Errorf("environment %q not found", env)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		// Step 4: deploy the build
		// (see also the "deployment" and "project" parent commands).
		deployment, err := sdktypes.DeploymentFromProto(&sdktypes.DeploymentPB{
			EnvId:   eid.String(),
			BuildId: bid.String(),
		})
		if err != nil {
			return fmt.Errorf("invalid deployment: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		dep := common.Client().Deployments()
		did, err := dep.Create(ctx, deployment)
		if err != nil {
			return fmt.Errorf("create deployment: %w", err)
		}
		logFunc(cmd, "exec")(fmt.Sprintf("create_deployment: created %q", did))

		// Step 5: activate the deployment
		// (see also the "deployment" parent command).
		if err := dep.Activate(ctx, did); err != nil {
			return fmt.Errorf("activate deployment: %w", err)
		}
		logFunc(cmd, "exec")("activate_deployment: activated")

		return nil
	},
})

func init() {
	// Command-specific flags.
	deployCmd.Flags().StringArrayVarP(&dirPaths, "dir", "d", []string{}, "0 or more directory paths (default = manifest directory)")
	deployCmd.Flags().StringArrayVarP(&filePaths, "file", "f", []string{}, "0 or more file paths")
	kittehs.Must0(deployCmd.MarkFlagDirname("dir"))
	kittehs.Must0(deployCmd.MarkFlagFilename("file"))
	deployCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "only show errors, if any")
	deployCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	deployCmd.Flags().StringVarP(&projectName, "project-name", "n", "", "project name")
}

func applyManifest(cmd *cobra.Command, args []string) (string, error) {
	// Read and parse the manifest file.
	data, path, err := common.Consume(args)
	if err != nil {
		return "", err
	}

	actions, err := plan(cmd, data, path, projectName)
	if err != nil {
		return "", err
	}

	client := common.Client()
	ctx, cancel := common.LimitedContext()
	defer cancel()

	// Execute the plan.
	pids, err := manifest.Execute(ctx, actions, client, logFunc(cmd, "exec"))
	if err != nil {
		return "", err
	}
	if len(pids) == 0 {
		// Execute didn't return a new project ID because the project already exists,
		// so get the project name from the manifest instead. It's safe to ignore the
		// error here because we already ran Read() successfully inside plan() above.
		m := kittehs.Must1(manifest.Read(data, path))
		return m.Project.Name, nil
	}
	if len(pids) > 1 {
		return "", fmt.Errorf("expected 1 project ID, got %d", len(pids))
	}

	return pids[0].String(), nil // Exactly one new project created.
}
