package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var (
	mode      string
	readyFile string
)

var upCmd = common.StandardCommand(&cobra.Command{
	Use:     "up [--mode={default|dev|test}] [--ready-file=FILE]",
	Short:   "Start local server",
	Aliases: []string{"u"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := svc.ParseMode(mode)
		if err != nil {
			return fmt.Errorf("mode: %w", err)
		}

		ctx := cmd.Root().Context()
		app, err := svc.New(common.Config(), svc.RunOptions{Mode: m})
		if err != nil {
			return fmt.Errorf("new service: %w", err)
		}

		if err := app.Start(ctx); err != nil {
			return fmt.Errorf("fx app start: %w", err)
		}

		if readyFile != "" {
			if err := os.WriteFile(readyFile, []byte("ready"), 0o644); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "write ready file: %v\n", err)
			}
		}

		<-app.Wait()

		fmt.Println() // End the output with "\n".

		return nil
	},
})

func init() {
	// Command-specific flags.
	upCmd.Flags().StringVarP(&mode, "mode", "m", "", "run mode: {default|dev|test}")

	upCmd.Flags().StringVarP(&readyFile, "ready-file", "r", "", "write a file when the server is ready")
	kittehs.Must0(upCmd.MarkFlagFilename("ready-file"))
}
