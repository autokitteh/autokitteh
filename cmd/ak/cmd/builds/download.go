package builds

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var downloadCmd = common.StandardCommand(&cobra.Command{
	Use:     "download <build ID> [--output=...]",
	Short:   "Download build data from server",
	Aliases: []string{"down", "dl", "do"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		b, id, err := r.BuildID(args[0])
		if err != nil {
			return err
		}
		if b == nil {
			err = fmt.Errorf("build ID %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		reader, err := builds().Download(ctx, id)
		if err != nil {
			return fmt.Errorf("download build: %w", err)
		}
		defer reader.Close()

		dst := os.Stdout
		if output != "-" {
			dst, err = outputFile()
			if err != nil {
				return err
			}
			defer dst.Close()
		}

		if _, err := io.Copy(dst, reader); err != nil {
			return fmt.Errorf("write output: %w", err)
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	downloadCmd.Flags().StringVarP(&output, "output", "o", defaultOutput, `output file path, or "-" for stdout`)
	kittehs.Must0(createCmd.MarkFlagFilename("output"))
}
