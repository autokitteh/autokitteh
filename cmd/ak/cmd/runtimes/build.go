package runtimes

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/tools/txtar"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	defaultOutput = "build.akb"
)

var (
	output   string
	dir      string
	values   []string
	describe bool
)

var buildCmd = common.StandardCommand(&cobra.Command{
	Use:     "build [<path> [<path> [...]] [--dir=...] [--output=...] [--values=...] [--describe]",
	Short:   `Build program and save it locally (see also "project build" command)`,
	Aliases: []string{"b"},
	Args:    cobra.ArbitraryArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		syms, err := kittehs.TransformError(values, sdktypes.ParseSymbol)
		if err != nil {
			return err
		}

		if dir == "" {
			dir = "."
		}

		var srcFS fs.FS = os.DirFS(dir)

		if txtarFile {
			if len(args) > 1 {
				return fmt.Errorf("txtar works with a single file or stdin")
			}

			var r io.Reader = os.Stdin
			if len(args) > 0 {
				f, err := os.Open(args[0])
				if err != nil {
					return fmt.Errorf("open: %w", err)
				}

				defer f.Close()

				r = f
			}

			bs, err := io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("read: %w", err)
			}

			a := txtar.Parse(bs)

			if srcFS, err = kittehs.TxtarToFS(a); err != nil {
				return fmt.Errorf("internal error: %w", err)
			}

			if srcFS, err = fs.Sub(srcFS, dir); err != nil {
				return err
			}

			args = nil
		}

		b, err := build(srcFS, args, syms)
		if err != nil {
			return err
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

	buildCmd.Flags().StringSliceVarP(&values, "values", "i", nil, "comma-separated input value names")

	buildCmd.Flags().StringVarP(&output, "output", "o", defaultOutput, `output file path, or "-" for stdout`)
	kittehs.Must0(buildCmd.MarkFlagFilename("output"))

	buildCmd.Flags().BoolVarP(&describe, "describe", "d", false, "describe build when completed")
	buildCmd.Flags().BoolVar(&txtarFile, "txtar", false, "input file is a txtar archive containing program")
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

func build(srcFS fs.FS, paths []string, syms []sdktypes.Symbol) (*sdkbuildfile.BuildFile, error) {
	ctx, cancel := common.LimitedContext()
	defer cancel()

	if len(paths) != 0 {
		files := make(map[string][]byte, len(paths))

		for _, path := range paths {
			data, err := fs.ReadFile(srcFS, filepath.Clean(path))
			if err != nil {
				return nil, fmt.Errorf("read file %q: %w", path, err)
			}

			files[path] = data
		}

		var err error
		if srcFS, err = kittehs.MapToMemFS(files); err != nil {
			return nil, fmt.Errorf("create memory filesystem: %w", err)
		}
	}

	b, err := runtimes().Build(ctx, srcFS, syms, nil)
	if err != nil {
		return nil, fmt.Errorf("create build: %w", err)
	}

	return b, nil
}
