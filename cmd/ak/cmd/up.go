package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/backend/basesvc"
	"go.autokitteh.dev/autokitteh/backend/configset"
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
		m, err := configset.ParseMode(mode)
		if err != nil {
			return fmt.Errorf("mode: %w", err)
		}

		ctx := cmd.Root().Context()
		app := svc.New(common.Config(), basesvc.RunOptions{Mode: m})
		if err := app.Start(ctx); err != nil {
			return fmt.Errorf("fx app start: %w", err)
		}

		<-app.Wait()
		fmt.Println() // End the output with "\n".

		if err := app.Stop(ctx); err != nil {
			return fmt.Errorf("fx app stop: %w", err)
		}
		return nil
	},
})

func init() {
	// Command-specific flags.
	upCmd.Flags().StringVarP(&mode, "mode", "m", "", "run mode: {default|dev|test}")
}
