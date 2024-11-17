package projects

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	projectsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1"
)

var (
	lintOpts struct {
		manifestPath string
	}
)

func init() {
	// Command-specific flags.
	lintCmd.Flags().StringVarP(&lintOpts.manifestPath, "manifest", "m", "", "YAML manifest file containing project settings")
}

var lintCmd = common.StandardCommand(&cobra.Command{
	Use:   "lint",
	Short: "Lint a project",
	Args:  cobra.NoArgs,
	RunE:  runLint,
})

func runLint(cmd *cobra.Command, args []string) error {
	file, err := os.Open(lintOpts.manifestPath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	resources := map[string][]byte{
		"autokitteh.yaml": data,
	}

	ctx, cancel := common.LimitedContext()
	defer cancel()
	vs, err := projects().Lint(ctx, sdktypes.InvalidProjectID, resources)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	for i := range vs { // Can't iterate over values, they contain lock.
		v := &vs[i]
		level := levelName(v.Level)
		fmt.Fprintf(w, "%s:%d - %s - %s\n", v.FileName, v.Line, level, v.Message)
	}

	if len(vs) > 0 {
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
