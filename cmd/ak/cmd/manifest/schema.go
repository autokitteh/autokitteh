package manifest

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/backend/apply"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var schemaCmd = common.StandardCommand(&cobra.Command{
	Use:     "schema",
	Short:   "Show YAML file schema",
	Aliases: []string{"s"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), apply.JSONSchemaString)
		return nil
	},
})
