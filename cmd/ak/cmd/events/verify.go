package events

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var verifyCmd = common.StandardCommand(&cobra.Command{
	Use:     "verify-filter <filter_expression>",
	Short:   "Verify if a CEL filter expression is valid",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"vf"},
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := args[0]
		if err := sdktypes.VerifyEventFilter(filter); err != nil {
			return fmt.Errorf("verify filter: %w", err)
		}
		common.RenderKV("result", "filter expression is valid")
		return nil
	},
})
