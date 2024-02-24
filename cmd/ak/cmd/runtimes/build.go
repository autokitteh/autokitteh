package runtimes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	defaultOutput = "build.akb"
)

var (
	output   string
	dir      string
	paths    []string
	values   []string
	describe bool
)

var buildCmd = common.StandardCommand(&cobra.Command{
	Use:     "build [--dir=...] <--path=...> [--output=...] [--describe]",
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
			files := make(map[string][]byte, len(paths))

			for _, path := range paths {
				data, err := fs.ReadFile(srcFS, filepath.Clean(path))
				if err != nil {
					return fmt.Errorf("read file %q: %w", path, err)
				}

				files[path] = data
			}

			if srcFS, err = kittehs.MapToMemFS(files); err != nil {
				return fmt.Errorf("create memory filesystem: %w", err)
			}
		}

		b, err := client().Build(ctx, srcFS, vns, nil)
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
	buildCmd.Flags().StringVarP(&dir, "dir", "w", ".", "working directory")
	kittehs.Must0(buildCmd.MarkFlagDirname("dir"))

	buildCmd.Flags().StringSliceVarP(&paths, "path", "p", nil, "one or more files to process")
	kittehs.Must0(buildCmd.MarkFlagFilename("path"))

	buildCmd.Flags().StringSliceVarP(&values, "values", "i", nil, "comma-separated input value names")

	buildCmd.Flags().StringVarP(&output, "output", "o", defaultOutput, `output file path, or "-" for stdout`)
	kittehs.Must0(buildCmd.MarkFlagFilename("output"))

	buildCmd.Flags().BoolVarP(&describe, "describe", "d", false, "describe build when completed")
}

func outputFile() (*os.File, error) {
	if output == "" {
		output = defaultOutput
	}

	f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	return f, nil
}
