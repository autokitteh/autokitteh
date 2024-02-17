package projects

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--fail]",
	Short:   "List all projects",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		ps, err := projects().List(ctx)
		if err != nil {
			return fmt.Errorf("list projects: %w", err)
		}

		if len(ps) == 0 {
			var dummy *sdktypes.Project
			return common.FailIfNotFound(cmd, "projects", dummy)
		}

		common.RenderList(ps)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)
}
