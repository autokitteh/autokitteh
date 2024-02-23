package builds

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--fail]",
	Short:   "List all uploaded project builds",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		bs, err := builds().List(ctx, sdkservices.ListBuildsFilter{})
		if err != nil {
			return fmt.Errorf("list builds: %w", err)
		}

		if len(bs) == 0 {
			return common.FailNotFound(cmd, "builds")
		}

		common.RenderList(bs)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)
}
