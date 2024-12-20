package projects

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var outputDirectory string

var downloadCmd = common.StandardCommand(&cobra.Command{
	Use:   "download <project name or ID> [--output-dir <path>] [--fail]",
	Short: "Download project resources",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		pid, err := r.ProjectNameOrID(ctx, args[0])
		if err = common.AddNotFoundErrIfCond(err, pid.IsValid()); err != nil {
			return common.WrapError(err, "project")
		}

		resources, err := projects().DownloadResources(ctx, pid)
		if err = common.AddNotFoundErrIfCond(err, len(resources) > 0); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "resources")
		}

		for filename, data := range resources {
			fulllPath := filepath.Join(outputDirectory, filename)

			if err := os.MkdirAll(path.Dir(fulllPath), 0o755); err != nil {
				return fmt.Errorf("create output directory: %w", err)
			}

			if err := os.WriteFile(fulllPath, data, 0o644); err != nil {
				return fmt.Errorf("write file: %w", err)
			}
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	downloadCmd.Flags().StringVarP(&outputDirectory, "output-dir", "o", ".", "path to output directory")

	common.AddFailIfNotFoundFlag(downloadCmd)
}
