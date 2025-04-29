package temporal

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/backend/temporaldevsrv"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

var cached temporaldevsrv.CachedDownload

var downloadCmd = common.StandardCommand(&cobra.Command{
	Use:   `download`,
	Short: "Download Temporal's dev server",
	Args:  cobra.NoArgs,

	RunE: func(*cobra.Command, []string) error {
		l, err := zap.NewDevelopment()
		if err != nil {
			return fmt.Errorf("logger: %w", err)
		}

		path, err := temporaldevsrv.Download(context.Background(), cached, l)

		common.RenderKVIfV("path", path)

		return err
	},
})

func init() {
	downloadCmd.Flags().StringVarP(&cached.Version, "version", "v", "", "desired version of the dev server")
	downloadCmd.Flags().StringVarP(&cached.DestDir, "dest-dir", "d", xdg.CacheHomeDir(), "destination directory for the dev server")
}
