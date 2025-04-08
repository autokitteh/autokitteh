package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	cmdmanifest "go.autokitteh.dev/autokitteh/cmd/ak/cmd/manifest"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	manifestPath, project, projectName, org string

	filePaths, dirPaths []string
)

var deployCmd = common.StandardCommand(&cobra.Command{
	Use:   "deploy {--manifest <file> [--project-name <name>]|--project <name or ID>} [--org org] [--dir <path> [...]] [--file <path> [...]] ",
	Short: "Create, configure, build, deploy, and activate project",
	Long:  `Create, configure, build, deploy, and activate project - see also the "manifest", "build", "deployment", and "project" parent commands`,
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		var oid sdktypes.OrgID
		if org != "" {
			var err error
			if oid, err = r.Org(ctx, org); err != nil {
				return fmt.Errorf("org: %w", err)
			}

		}

		// Step 1: apply the manifest file, if provided
		// (see also the "manifest" parent command).
		if manifestPath != "" {
			var err error
			project, err = applyManifest(cmd, manifestPath, projectName, oid)
			if err != nil {
				return err
			}
		} else if projectName != "" {
			return errors.New("project name provided without manifest")
		}

		pid, err := r.ProjectNameOrID(ctx, oid, project)
		if err != nil {
			err = fmt.Errorf("project: %w", err)

			if errors.Is(err, sdkerrors.ErrNotFound) {
				err = common.NewExitCodeError(common.NotFoundExitCode, err)
			}

			return err
		}

		if pid.IsValid() {
			logFunc(cmd, "plan")(fmt.Sprintf("project %q: found, id=%q", project, pid))
		}

		// Step 2: build the project (see also the "build" and "project" parent commands).
		if len(dirPaths) == 0 && len(filePaths) == 0 {
			if manifestPath == "" {
				return errors.New("no dir/file paths provided")
			}
			dirPaths = append(dirPaths, filepath.Dir(manifestPath))
		}

		bid, err := common.BuildProject(pid, dirPaths, filePaths)
		if err != nil {
			return err
		}
		logFunc(cmd, "exec")(fmt.Sprintf("create_build: created %q", bid))

		// Step 3: deploy the build
		// (see also the "deployment" and "project" parent commands).
		deployment, err := sdktypes.DeploymentFromProto(&sdktypes.DeploymentPB{
			ProjectId: pid.String(),
			BuildId:   bid.String(),
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
	deployCmd.Flags().StringVarP(&org, "org", "o", "", "org to use for manifest")
	deployCmd.Flags().StringVarP(&project, "project", "p", "", "existing project name or ID")
	deployCmd.MarkFlagsOneRequired("manifest", "project")
	deployCmd.MarkFlagsMutuallyExclusive("manifest", "project")
	deployCmd.MarkFlagsMutuallyExclusive("project-name", "project")

	deployCmd.Flags().StringArrayVarP(&dirPaths, "dir", "d", []string{}, "0 or more directory paths (default = manifest directory)")
	deployCmd.Flags().StringArrayVarP(&filePaths, "file", "f", []string{}, "0 or more file paths")
	kittehs.Must0(deployCmd.MarkFlagDirname("dir"))
	kittehs.Must0(deployCmd.MarkFlagFilename("file"))
}

func applyManifest(cmd *cobra.Command, manifestPath, projectName string, oid sdktypes.OrgID) (string, error) {
	// Read and parse the manifest file.
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", common.NewExitCodeError(common.NotFoundExitCode, err)
	}

	m, err := manifest.Read(data)
	if err != nil {
		return "", err
	}

	client := common.Client()
	ctx, cancel := common.LimitedContext()
	defer cancel()

	// Plan the actions to execute.
	actions, err := manifest.Plan(
		ctx,
		m,
		client,
		manifest.WithLogger(logFunc(cmd, "plan")),
		manifest.WithProjectName(projectName),
		manifest.WithOrgID(oid),
	)
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
