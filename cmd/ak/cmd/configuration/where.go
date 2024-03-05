package configuration

import (
	"fmt"

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
		fmt.Println("Config home directory:", xdg.ConfigHomeDir())
		fmt.Println("Data home directory:  ", xdg.DataHomeDir())
		fmt.Println()
		fmt.Println("Override environment variable names:")
		fmt.Println(xdg.ConfigEnvVar)
		fmt.Println(xdg.DataEnvVar)

		return nil
	},
})
