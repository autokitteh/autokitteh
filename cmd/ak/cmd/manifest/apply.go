package manifest

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"

	"go.autokitteh.dev/autokitteh/backend/apply"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

type output struct {
	Logs       any `json:"logs"`
	Operations any `json:"operations"`
}

var dryRun, noValidate bool

var applyCmd = common.StandardCommand(&cobra.Command{
	Use:     "apply [file] [--no-validate] [--dry-run]",
	Short:   "Apply project configuration from file or stdin",
	Aliases: []string{"a"},
	Args:    cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			path string
			data []byte
			err  error
		)
		if len(args) == 0 {
			data, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("stdin: %w", err)
			}
		} else {
			path = args[0]
			data, err = os.ReadFile(path)
			if err != nil {
				return err
			}
		}

		var root apply.Root
		if err := yaml.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("invalid YAML input: %w", err)
		}

		if !noValidate {
			res, err := gojsonschema.Validate(
				gojsonschema.NewStringLoader(apply.JSONSchemaString),
				gojsonschema.NewGoLoader(&root),
			)
			if err != nil {
				return fmt.Errorf("validate: %w", err)
			}
			if !res.Valid() {
				msg := strings.Join(kittehs.Transform(res.Errors(), err2str), "\n")
				return fmt.Errorf("invalid YAML semantics: %s", msg)
			}
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		a := apply.Applicator{Svcs: common.Client(), Path: path}
		if err := a.Plan(ctx, &root); err != nil {
			return err
		}

		common.RenderKV("plan", output{
			Logs:       a.Logs(),
			Operations: a.Operations(),
		})

		if dryRun {
			return nil
		}

		a.ResetLogs()
		err = a.Apply(ctx)
		common.RenderKV("apply_logs", a.Logs())
		return err
	},
})

func init() {
	// Command-specific flags.
	applyCmd.Flags().BoolVarP(&noValidate, "no-validate", "n", false, "don't validate before applying")
	applyCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "only show plan, don't apply")
}

func err2str(err gojsonschema.ResultError) string {
	return fmt.Sprintf("- %s: %s", err.Field(), err.Description())
}
