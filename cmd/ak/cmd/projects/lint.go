package projects

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
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

func getManifest(resources map[string][]byte, manifestFile string) (*manifest.Manifest, error) {
	data := resources[manifestFile]
	if data == nil {
		return nil, nil
	}

	return manifest.Read(data, manifestFile)
}

func findProjectNameOrID(projectNameOrID string, projectDir string, m *manifest.Manifest) (string, error) {
	if projectNameOrID != "" {
		return projectNameOrID, nil
	}

	// Try pid file
	pidFile := path.Join(projectDir, ".autokitteh", "pid")
	data, err := os.ReadFile(pidFile)
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	if m != nil {
		return m.Project.Name, nil
	}

	return "", fmt.Errorf("can't determine project name or ID")
}

func printViolation(w io.Writer, v *projectsv1.CheckViolation) {
	// TODO: JSON?
	level := levelName(v.Level)
	// FIXME (ENG-1867): RuleId arrives as empty string.
	fmt.Fprintf(w, "%s:%d - %s - %s\n", v.FileName, v.Line, level, v.Message)
}

func runLint(cmd *cobra.Command, args []string) error {
	r := resolver.Resolver{Client: common.Client()}
	ctx, cancel := common.LimitedContext()
	defer cancel()

	resources, err := buildResources()
	if err != nil {
		return err
	}

	manifestFile := path.Base(lintOpts.manifestPath)
	m, err := getManifest(resources, manifestFile)
	if err != nil {
		return err
	}

	projectDir := path.Base(lintOpts.manifestPath)
	projectNameOrID, err := findProjectNameOrID(lintOpts.projectNameOrID, projectDir, m)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	_, projectID, err := r.ProjectNameOrID(ctx, projectNameOrID)
	switch err {
	case sdkerrors.ErrNotFound: // new project
		// no need to check
	case nil: // Existing project
		if m != nil { // Check that manifest is not outdated
			actions, err := manifest.Plan(ctx, m, common.Client(), manifest.WithLogger(emptyLog))
			if err != nil {
				return err
			}
			if len(actions) > 0 {
				v := projectsv1.CheckViolation{
					FileName: manifestFile,
					Level:    projectsv1.CheckViolation_LEVEL_WARNING,
					Message:  "outdated manifest",
				}
				printViolation(w, &v)
			}
		}
	}

	vs, err := projects().Lint(ctx, projectID, resources, manifestFile)
	if err != nil {
		return err
	}

	ok := true
	for _, v := range vs {
		printViolation(w, v)
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

func emptyLog(string) {}
