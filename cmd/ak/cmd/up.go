package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var mode string

var upCmd = common.StandardCommand(&cobra.Command{
	Use:     "up [--mode={default|dev|test}]",
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

		<-app.Wait()

		fmt.Println() // End the output with "\n".

		return nil
	},
})

func init() {
	// Command-specific flags.
	upCmd.Flags().StringVarP(&mode, "mode", "m", "", "run mode: {default|dev|test}")
}
