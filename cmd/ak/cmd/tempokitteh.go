package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/tempokittehsvc"
	"go.autokitteh.dev/autokitteh/runtimes/configrt"
	"go.autokitteh.dev/autokitteh/runtimes/pythonrt"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/tempokitteh"
)

const addr = "0.0.0.0:9988"

var dir, qname string

var tempokittehCmd = common.StandardCommand(&cobra.Command{
	Use:   "tempokitteh [<path> [<path> [...]] [--dir=...] --task-queue-name=<name>",
	Short: "Run temporal worker",
	Args:  cobra.ArbitraryArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		m, err := common.ParseModeFlag()
		if err != nil {
			return err
		}

		ropts := tempokittehsvc.RunOptions{
			Mode:        m,
			TKQueueName: qname,
		}

		var (
			l        *zap.Logger
			runtimes sdkservices.Runtimes
			calls    sessioncalls.Calls
			w        tempokitteh.Worker
		)

		opts := append(
			tempokittehsvc.NewOpts(common.Config(), ropts),
			fx.Populate(&l, &runtimes, &w, &calls),
		)

		app := fx.New(opts...)
		if err := app.Err(); err != nil {
			return err
		}

		if dir == "" {
			dir = "."
		}

		b, err := common.Build(runtimes, os.DirFS(dir), args, nil)
		if err != nil {
			return fmt.Errorf("build: %w", err)
		}

		if err := tempokitteh.Register(ctx, l.Named("tk"), w, runtimes, calls, b); err != nil {
			return fmt.Errorf("register: %w", err)
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
	tempokittehCmd.Flags().StringVarP(&dir, "dir", "d", "", "root directory")
	tempokittehCmd.Flags().StringVarP(&qname, "task-queue-name", "q", "", "task queue name")
	tempokittehCmd.MarkFlagRequired("task-queue-name")
	common.AddModeFlag(tempokittehCmd)
	common.AddSilentFlag(tempokittehCmd)
}

func initRuntimes(l *zap.Logger, mux *http.ServeMux) (sdkservices.Runtimes, error) {
	pyrt, err := pythonrt.New(pythonrt.Configs.Dev, l, func() string { return addr })
	if err != nil {
		return nil, fmt.Errorf("python: %w", err)
	}

	pythonrt.ConfigureWorkerGRPCHandler(l, mux)

	return sdkruntimes.New([]*sdkruntimes.Runtime{
		starlarkrt.New(),
		configrt.New(),
		pyrt,
	})
}
