package envs

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <environment name or ID> [--project=...] [--fail]",
	Short:   "Get execution environment details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		e, _, err := r.EnvNameOrID(args[0], project)
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "env", e.IsValid()); err != nil {
			return err
		}

		common.RenderKVIfV("env", e)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
