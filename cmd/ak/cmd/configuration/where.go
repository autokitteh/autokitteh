package configuration

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

var whereCmd = common.StandardCommand(&cobra.Command{
	Use:     "where",
	Short:   "Where are the config and data directories",
	Aliases: []string{"w"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := xdg.ConfigHomeDir()
		if strings.Contains(cfg, " ") {
			cfg = `"` + cfg + `"`
		}

		data := xdg.DataHomeDir()
		if strings.Contains(data, " ") {
			data = `"` + data + `"`
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
