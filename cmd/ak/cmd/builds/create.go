package builds

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuild"
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

		url, err := url.Parse(dir)
		if err != nil {
			return fmt.Errorf("invalid root dir: %w", err)
		}

		b, err := sdkbuild.Build(ctx, common.Client().Runtimes(), url, paths, vns, nil)
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
	kittehs.Must0(createCmd.MarkFlagRequired("path"))

	createCmd.Flags().StringSliceVarP(&values, "values", "i", nil, "comma-separated input value names")

	createCmd.Flags().StringVarP(&output, "output", "o", defaultOutput, `output file path, or "-" for stdout`)
	kittehs.Must0(createCmd.MarkFlagFilename("output"))

	createCmd.Flags().BoolVarP(&describe, "describe", "d", false, "describe build when completed")
}
