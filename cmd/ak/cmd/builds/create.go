package builds

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/psanford/memfs"
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/backend/runtimes"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	dir      string
	paths    []string
	values   []string
	describe bool
)

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create [--dir=...] <--path=...> [--output=...] [--describe]",
	Short:   `Build program and save it locally (see also "project build" command)`,
	Aliases: []string{"c"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		vns, err := kittehs.TransformError(values, sdktypes.ParseSymbol)
		if err != nil {
			return fmt.Errorf("invalid values: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		srcFS := os.DirFS(dir)

		if len(paths) != 0 {
			memfs := memfs.New()

			for _, path := range paths {
				data, err := fs.ReadFile(srcFS, path)
				if err != nil {
					return fmt.Errorf("read file %q: %w", path, err)
				}

				if err := memfs.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					return fmt.Errorf("create directory %q in memory: %w", path, err)
				}

				if err := memfs.WriteFile(path, data, 0o600); err != nil {
					return fmt.Errorf("write file %q into memory: %w", path, err)
				}
			}

			srcFS = memfs
		}

		// Currently uses only local runtimes - no RPC support yet.
		b, err := sdkruntimes.Build(ctx, runtimes.New(), srcFS, vns, nil)
		if err != nil {
			return fmt.Errorf("create build: %w", err)
		}

		dst := os.Stdout
		if output != "-" {
			dst, err = outputFile()
			if err != nil {
				return err
			}
			defer dst.Close()
		}

		if err := b.Write(dst); err != nil {
			return fmt.Errorf("write output: %w", err)
		}

		if describe {
			b.OmitContent()
			common.Render(b)
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	createCmd.Flags().StringVarP(&dir, "dir", "w", ".", "working directory")
	kittehs.Must0(createCmd.MarkFlagDirname("dir"))

	createCmd.Flags().StringSliceVarP(&paths, "path", "p", nil, "one or more files to process")
	kittehs.Must0(createCmd.MarkFlagFilename("path"))

	createCmd.Flags().StringSliceVarP(&values, "values", "i", nil, "comma-separated input value names")

	createCmd.Flags().StringVarP(&output, "output", "o", defaultOutput, `output file path, or "-" for stdout`)
	kittehs.Must0(createCmd.MarkFlagFilename("output"))

	createCmd.Flags().BoolVarP(&describe, "describe", "d", false, "describe build when completed")
}
