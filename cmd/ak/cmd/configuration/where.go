package configuration

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/backend/config"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var whereCmd = common.StandardCommand(&cobra.Command{
	Use:     "where",
	Short:   "Where are the config and data directories",
	Aliases: []string{"w"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Config home directory:", config.ConfigHomeDir())
		fmt.Println("Data home directory:  ", config.DataHomeDir())
		fmt.Println()
		fmt.Println("Override environment variable names:")
		fmt.Println(config.ConfigEnvVar)
		fmt.Println(config.DataEnvVar)

		return nil
	},
})
