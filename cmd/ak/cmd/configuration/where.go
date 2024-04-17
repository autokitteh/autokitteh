package configuration

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

type config struct {
	ConfigDir        string `json:"config_dir"`
	DataDir          string `json:"data_dir"`
	ConfigEnvVarName string `json:"config_env_var_name"`
	DataEnvVarName   string `json:"data_env_var_name"`
}

func (c config) Text() string {
	cfg := c.ConfigDir
	if strings.Contains(cfg, " ") {
		cfg = strconv.Quote(cfg)
	}

	data := c.DataDir
	if strings.Contains(data, " ") {
		data = strconv.Quote(data)
	}

	var out strings.Builder

	fmt.Fprintln(&out, "Config home directory:", cfg)
	fmt.Fprintln(&out, "Data home directory:  ", data)
	fmt.Fprintln(&out, "")
	fmt.Fprintln(&out, "Override environment variable names:")
	fmt.Fprintln(&out, c.ConfigEnvVarName)
	fmt.Fprintln(&out, c.DataEnvVarName)

	return out.String()
}

var whereCmd = common.StandardCommand(&cobra.Command{
	Use:     "where",
	Short:   "Where are the config and data directories",
	Aliases: []string{"w"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		c := config{
			ConfigDir:        xdg.ConfigHomeDir(),
			DataDir:          xdg.DataHomeDir(),
			ConfigEnvVarName: xdg.ConfigEnvVar,
			DataEnvVarName:   xdg.DataEnvVar,
		}

		common.Render(c)

		return nil
	},
})
