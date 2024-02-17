package mappings

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <mapping ID> [--fail]",
	Short:   "Get connection mapping details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		m, _, err := r.MappingID(args[0])
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "mapping", m); err != nil {
			return err
		}

		common.RenderKVIfV("mapping", m)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
