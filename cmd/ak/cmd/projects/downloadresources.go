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

var downloadResourcesCmd = common.StandardCommand(&cobra.Command{
	Use:   "download-resources <project name or ID> [--output-dir=...]",
	Short: "Download the project's resources",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		_, pid, err := r.ProjectNameOrID(args[0])
		if err != nil {
			return err
		}

		if !pid.IsValid() {
			return fmt.Errorf("project %s not found", args[0])
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		resources, err := projects().DownloadResources(ctx, pid)
		if err != nil {
			return fmt.Errorf("download resources: %w", err)
		}

		if len(resources) == 0 {
			fmt.Println("no resources found")
			return nil
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
	downloadResourcesCmd.Flags().StringVarP(&outputDirectory, "output-dir", "o", ".", "path to output directory")
}
