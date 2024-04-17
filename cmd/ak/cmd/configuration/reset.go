package configuration

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

var resetData, resetCfg bool

var resetCmd = common.StandardCommand(&cobra.Command{
	Use:   "reset [--data] [--config]",
	Short: "Reset persistent configuration",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, data := xdg.ConfigHomeDir(), xdg.DataHomeDir()

		if cfg == data && (!resetData || !resetCfg) {
			return errors.New("data and config dirs are the same, must specify both --data and --config")
		}

		if resetData {
			return os.RemoveAll(filepath.Join(data, "*"))
		}

		if resetCfg {
			return os.RemoveAll(filepath.Join(cfg, "*"))
		}

		return nil
	},
})

func init() {
	resetCmd.Flags().BoolVarP(&resetData, "data", "d", false, "reset data configuration")
	resetCmd.Flags().BoolVarP(&resetCfg, "config", "c", false, "reset configuration")
	resetCmd.MarkFlagsOneRequired("data", "config")
}
