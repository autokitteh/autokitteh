package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/backend/basesvc"
	"go.autokitteh.dev/autokitteh/backend/configset"
	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var mode string

var App *fx.App

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
		App = svc.New(common.Config(), basesvc.RunOptions{Mode: m})
		if err := App.Start(ctx); err != nil {
			return fmt.Errorf("fx App start: %w", err)
		}

		select {
		case sig := <-App.Done():
			fmt.Fprintf(os.Stderr, "got shutdown signal: %v\n", sig)
		case <-ctx.Done():
			fmt.Fprintf(os.Stderr, "shutting down...\n")
			if err := App.Stop(context.Background()); err != nil {
				return fmt.Errorf("fx app stop: %w", err)
			}
			<-App.Wait()
		}

		fmt.Fprintln(os.Stderr) // End the output with "\n".
		return nil
	},
})

func init() {
	// Command-specific flags.
	upCmd.Flags().StringVarP(&mode, "mode", "m", "", "run mode: {default|dev|test}")
}
