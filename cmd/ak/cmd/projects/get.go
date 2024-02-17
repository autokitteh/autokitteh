package projects

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <project name or ID> [--fail]",
	Short:   "Get project details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		p, _, err := r.ProjectNameOrID(args[0])
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "project", p); err != nil {
			return err
		}

		common.RenderKVIfV("project", p)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
