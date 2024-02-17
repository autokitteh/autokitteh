package projects

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var (
	sourcePath string
)

var setResourcesCmd = common.StandardCommand(&cobra.Command{
	Use:   "set-resources <project name or ID>  <--path=...>",
	Short: "Set the project's resources",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		_, pid, err := r.ProjectNameOrID(args[0])
		if err != nil {
			return err
		}

		if pid == nil {
			return fmt.Errorf("project %s not found", args[0])
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		uploads := map[string][]byte{}

		if isFile(sourcePath) {
			data, err := os.ReadFile(sourcePath)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			uploads[filepath.Base(sourcePath)] = data
		} else {
			if err := filepath.WalkDir(sourcePath, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if !d.IsDir() {
					data, err := os.ReadFile(path)
					if err != nil {
						return fmt.Errorf("read file: %w", err)
					}

					r, err := filepath.Rel(sourcePath, path)
					if err != nil {
						return fmt.Errorf("relative path: %w", err)
					}

					uploads[r] = data
				}
				return nil
			}); err != nil {
				return fmt.Errorf("loading content: %w", err)
			}
		}

		if err := projects().SetResources(ctx, pid, uploads); err != nil {
			return fmt.Errorf("set resources: %w", err)
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	setResourcesCmd.Flags().StringVarP(&sourcePath, "path", "p", "", "path to directory or file")

	kittehs.Must0(setResourcesCmd.MarkFlagRequired("path"))
}

func isFile(name string) bool {
	if fi, err := os.Stat(name); err == nil {
		if fi.Mode().IsRegular() {
			return true
		}
	}
	return false
}
