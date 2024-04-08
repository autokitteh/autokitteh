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
			data = fmt.Sprintf("%q", data)
			data = `"` + data + `"`
		}

		fmt.Println("Config home directory:", cfg)
		fmt.Println("Data home directory:  ", data)
		fmt.Println()
		fmt.Println("Override environment variable names:")
		fmt.Println(xdg.ConfigEnvVar)
		fmt.Println(xdg.DataEnvVar)

		return nil
	},
})
