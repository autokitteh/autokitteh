package configuration

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

var onlyCfg, onlyData bool

var whereCmd = common.StandardCommand(&cobra.Command{
	Use:     "where",
	Short:   "Where are the config and data directories",
	Aliases: []string{"w"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := xdg.ConfigHomeDir()

		if onlyCfg {
			fmt.Fprintln(cmd.OutOrStdout(), cfg)
			return nil
		}

		if strings.Contains(cfg, " ") {
			cfg = strconv.Quote(cfg)
		}

		data := xdg.DataHomeDir()

		if onlyData {
			fmt.Fprintln(cmd.OutOrStdout(), data)
			return nil
		}

		if strings.Contains(data, " ") {
			data = strconv.Quote(data)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Config home directory:", cfg)
		fmt.Fprintln(cmd.OutOrStdout(), "Data home directory:  ", data)
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "Override environment variable names:")
		fmt.Fprintln(cmd.OutOrStdout(), xdg.ConfigEnvVar)
		fmt.Fprintln(cmd.OutOrStdout(), xdg.DataEnvVar)

		return nil
	},
})

func init() {
	// Command-specific flags.
	whereCmd.Flags().BoolVarP(&onlyCfg, "cfg", "c", false, "print only config directory path")
	whereCmd.Flags().BoolVarP(&onlyData, "data", "d", false, "print only data directory path")
	whereCmd.MarkFlagsMutuallyExclusive("cfg", "data")
}
