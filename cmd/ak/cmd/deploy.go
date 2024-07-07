package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	cmdmanifest "go.autokitteh.dev/autokitteh/cmd/ak/cmd/manifest"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	manifestPath, project, env, projectName string

	filePaths, dirPaths []string
)

var deployCmd = common.StandardCommand(&cobra.Command{
	Use:   "deploy {--manifest <file> [--project-name <name>]|--project <name or ID>} [--dir <path> [...]] [--file <path> [...]] [--env <name or ID>]",
	Short: "Create, configure, build, deploy, and activate project",
	Long:  `Create, configure, build, deploy, and activate project - see also the "manifest", "build", "deployment", and "project" parent commands`,
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		// Step 1: apply the manifest file, if provided
		// (see also the "manifest" parent command).
		if manifestPath != "" {
			var err error
			project, err = applyManifest(cmd, manifestPath, projectName)
			if err != nil {
				return err
			}
		} else {
			if projectName != "" {
				return fmt.Errorf("project name provided without manifest")
			}

			p, pid, _ := r.ProjectNameOrID(ctx, project)
			if p.IsValid() {
				logFunc(cmd, "plan")(fmt.Sprintf("project %q: found, id=%q", project, pid))
			}
		}

		// Step 2: build the project (see also the "build" and "project" parent commands).
		if len(dirPaths) == 0 && len(filePaths) == 0 {
			if manifestPath == "" {
				return fmt.Errorf("no dir/file paths provided")
			}
			dirPaths = append(dirPaths, filepath.Dir(manifestPath))
		}

		bid, err := common.BuildProject(project, dirPaths, filePaths)
		if err != nil {
			return err
		}
		logFunc(cmd, "exec")(fmt.Sprintf("create_build: created %q", bid))

		// Step 3: parse the optional environment argument.
		e, eid, err := r.EnvNameOrID(ctx, env, project)
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

		dep := common.Client().Deployments()
		did, err := dep.Create(ctx, deployment)
		if err != nil {
			return fmt.Errorf("create deployment: %w", err)
		}
		logFunc(cmd, "exec")(fmt.Sprintf("create_deployment: created %q", did))

		// Step 5: activate the deployment (see also the "deployment" parent command).
		if err := dep.Activate(ctx, did); err != nil {
			return fmt.Errorf("activate deployment: %w", err)
		}
		logFunc(cmd, "exec")("activate_deployment: activated")

		return nil
	},
})

func init() {
	// Command-specific flags.
	deployCmd.Flags().StringVarP(&manifestPath, "manifest", "m", "", "YAML manifest file containing project settings")
	deployCmd.Flags().StringVarP(&projectName, "project-name", "n", "", "project name to use for manifest")
	deployCmd.Flags().StringVarP(&project, "project", "p", "", "existing project name or ID")
	deployCmd.MarkFlagsOneRequired("manifest", "project")
	deployCmd.MarkFlagsMutuallyExclusive("manifest", "project")

	deployCmd.Flags().StringArrayVarP(&dirPaths, "dir", "d", []string{}, "0 or more directory paths (default = manifest directory)")
	deployCmd.Flags().StringArrayVarP(&filePaths, "file", "f", []string{}, "0 or more file paths")
	kittehs.Must0(deployCmd.MarkFlagDirname("dir"))
	kittehs.Must0(deployCmd.MarkFlagFilename("file"))

	deployCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
}

func applyManifest(cmd *cobra.Command, manifestPath, projectName string) (string, error) {
	// Read and parse the manifest file.
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", common.NewExitCodeError(common.NotFoundExitCode, err)
	}

	m, err := manifest.Read(data, manifestPath)
	if err != nil {
		return "", err
	}

	client := common.Client()
	ctx, cancel := common.LimitedContext()
	defer cancel()

	// Plan the actions to execute.
	actions, err := manifest.Plan(ctx, m, client, manifest.WithLogger(logFunc(cmd, "plan")), manifest.WithProjectName(projectName))
	if err != nil {
		return "", err
	}

	// Execute the plan.
	effects, err := manifest.Execute(ctx, actions, client, logFunc(cmd, "exec"))
	if err != nil {
		return "", err
	}

	cmdmanifest.PrintSuggestions(cmd, effects)

	pids := effects.ProjectIDs()

	if len(pids) == 0 {
		return m.Project.Name, nil // Project already exists.
	}
	if len(pids) > 1 {
		return "", fmt.Errorf("expected 1 project ID, got %d", len(pids))
	}

	return pids[0].String(), nil // Exactly one new project created.
}

func logFunc(cmd *cobra.Command, prefix string) func(string) {
	return func(msg string) {
		fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", prefix, msg)
	}
}
