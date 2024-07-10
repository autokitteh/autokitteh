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
		ctx, cancel := common.LimitedContext()
		defer cancel()

		p, _, err := r.ProjectNameOrID(ctx, args[0])
		err = common.AddNotFoundErrIfCond(err, p.IsValid())
		if err = common.FailIfError2(cmd, err, "project"); err != nil {
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
