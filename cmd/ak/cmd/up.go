package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/backend/basesvc"
	"go.autokitteh.dev/autokitteh/backend/configset"
	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var (
	daemon bool
	mode   string
)

var upCmd = common.StandardCommand(&cobra.Command{
	Use:     "up [--mode={default|dev|test}] [--daemon]",
	Short:   "Start local server",
	Aliases: []string{"u"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		if daemon {
			// TODO: Implement. Need to detach only after all inits are done?
			return errors.New("not implemented yet")
		}

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

func RemoveUpCmd() { RootCmd.RemoveCommand(upCmd) }

func init() {
	// Command-specific flags.
	upCmd.Flags().StringVarP(&mode, "mode", "m", "", "run mode: {default|dev|test}")
	upCmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "run as daemon?")
}
