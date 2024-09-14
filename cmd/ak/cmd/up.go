package cmd

import (
	"fmt"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var upCmd = common.StandardCommand(&cobra.Command{
	Use:   "up [--mode {default|dev|test}]",
	Short: "Start local server",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := maxprocs.Set()
		if err != nil {
			return kittehs.ErrorWithPrefix("maxprocs set", err)
		}

		ctx := cmd.Root().Context()

		app, err := common.NewSvc(false)
		if err != nil {
			return kittehs.ErrorWithPrefix("new service", err)
		}

		if err := app.Start(ctx); err != nil {
			return kittehs.ErrorWithPrefix("fx app start", err)
		}

		<-app.Wait()

		fmt.Println() // End the output with "\n".
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddModeFlag(upCmd)
}
