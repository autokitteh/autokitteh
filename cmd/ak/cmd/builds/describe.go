package builds

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var describeCmd = common.StandardCommand(&cobra.Command{
	Use:   "describe <build ID> [--fail]",
	Short: "Describe build file stored in the server",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		_, bid, err := r.BuildID(ctx, args[0])
		if err != nil {
			return err
		}

		bf, err := builds().Describe(ctx, bid)
		if err != nil {
			return fmt.Errorf("download build: %w", err)
		}

		common.RenderKVIfV("build", bf)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(describeCmd)
}
