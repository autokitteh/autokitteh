package manifest

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/manifest"
)

var schemaCmd = common.StandardCommand(&cobra.Command{
	Use:     "schema",
	Short:   "Show YAML file schema",
	Aliases: []string{"s"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), manifest.JSONSchemaString)
		return nil
	},
})
