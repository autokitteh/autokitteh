package projects

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	projectsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1"
)

var (
	lintOpts struct {
		manifestPath    string
		projectNameOrID string
		dirPaths        []string
		filePaths       []string
	}
)

func init() {
	// Command-specific flags.
	lintCmd.Flags().StringVarP(&lintOpts.manifestPath, "manifest", "m", "", "YAML manifest file containing project settings")
	lintCmd.Flags().StringVarP(&lintOpts.projectNameOrID, "project", "p", "", "project name (or ID)")
	lintCmd.Flags().StringArrayVarP(&lintOpts.dirPaths, "dir", "d", []string{}, "0 or more directory paths")
	lintCmd.Flags().StringArrayVarP(&lintOpts.filePaths, "file", "f", []string{}, "0 or more file paths")
}

var lintCmd = common.StandardCommand(&cobra.Command{
	Use:   "lint",
	Short: "Lint a project",
	Args:  cobra.NoArgs,
	RunE:  runLint,
})

func buildResources() (map[string][]byte, error) {
	dirPaths := lintOpts.dirPaths
	if len(lintOpts.dirPaths)+len(lintOpts.filePaths) == 0 {
		dirPaths = []string{"."}
	}

	resources, err := common.CollectUploads(dirPaths, lintOpts.filePaths)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func findProjectID(ctx context.Context, r resolver.Resolver, manifestFile string) sdktypes.ProjectID {
	// command line first
	if lintOpts.projectNameOrID != "" {
		_, pid, err := r.ProjectNameOrID(ctx, lintOpts.projectNameOrID)
		if err == nil {
			return pid
		}
	}

	// pid file
	data, err := os.ReadFile(".autokitteh/pid")
	if err == nil {
		_, pid, err := r.ProjectNameOrID(ctx, string(data))
		if err == nil {
			return pid
		}
	}

	// Read from manifest
	data, err = os.ReadFile(manifestFile)
	if err == nil {
		m, err := manifest.Read(data, manifestFile)
		if err != nil && m.Project.Name != "" {
			_, pid, err := r.ProjectNameOrID(ctx, string(data))
			if err == nil {
				return pid
			}
		}
	}

	return sdktypes.InvalidProjectID
}

func runLint(cmd *cobra.Command, args []string) error {
	r := resolver.Resolver{Client: common.Client()}
	ctx, cancel := common.LimitedContext()
	defer cancel()

	projectID := findProjectID(ctx, r, lintOpts.manifestPath)

	resources, err := buildResources()
	if err != nil {
		return err
	}

	vs, err := projects().Lint(ctx, projectID, resources, lintOpts.manifestPath)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()

	ok := true
	for _, v := range vs {
		level := levelName(v.Level)
		// TODO: JSON?
		// FIXME (ENG-1867): RuleId arrives as empty string.
		fmt.Fprintf(w, "%s:%d - %s - %s\n", v.FileName, v.Line, level, v.Message)

		if v.Level == sdktypes.ViolationError {
			ok = false
		}
	}

	if !ok {
		return fmt.Errorf("lint errors")
	}

	return nil
}

func levelName(level projectsv1.CheckViolation_Level) string {
	name := projectsv1.CheckViolation_Level_name[int32(level)]
	if name == "" {
		return "UNKNOWN"
	}

	const prefix = "LEVEL_"
	return name[len(prefix):]
}
