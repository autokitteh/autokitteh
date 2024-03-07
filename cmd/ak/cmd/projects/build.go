package projects

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var (
	paths   []string
	uploads map[string][]byte
)

var buildCmd = common.StandardCommand(&cobra.Command{
	Use:   "build <project name or ID> --from <file or directory> [--from ...]",
	Short: `Build project (see also the "build" subcommands)`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		p, pid, err := r.ProjectNameOrID(args[0])
		if err != nil {
			return err
		}
		if !p.IsValid() {
			err = errors.New("project not found")
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		uploads = make(map[string][]byte)
		for _, path := range paths {
			fi, err := os.Stat(path)
			if err != nil {
				return err
			}

			// Upload an entire directory tree.
			if fi.IsDir() {
				err := filepath.WalkDir(path, walk(path))
				if err != nil {
					return err
				}
				continue
			}

			// Upload a single file.
			contents, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			uploads[fi.Name()] = contents
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		// Communicate with the server in 2 steps.
		if err := projects().SetResources(ctx, pid, uploads); err != nil {
			return fmt.Errorf("set resources: %w", err)
		}

		bid, err := projects().Build(ctx, pid)
		if err != nil {
			return fmt.Errorf("build project: %w", err)
		}

		common.RenderKVIfV("build_id", bid)
		return nil
	},
})

func init() {
	// Command-specific flags.
	buildCmd.Flags().StringArrayVarP(&paths, "from", "f", []string{}, "1 or more file or directory paths")
	kittehs.Must0(buildCmd.MarkFlagRequired("from"))
}

func walk(basePath string) fs.WalkDirFunc {
	return func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err // Abort the entire walk.
		}
		if d.IsDir() {
			return nil // Skip directory analysis, focus on files.
		}

		// Upload a single file, relative to the base path.
		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		uploads[relPath] = contents
		return nil
	}
}
