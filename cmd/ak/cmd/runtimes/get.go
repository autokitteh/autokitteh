package runtimes

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <runtime name> [--fail]",
	Short:   "Get runtime engine details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := sdktypes.ParseSymbol(args[0])
		if err != nil {
			return fmt.Errorf("name: %w", err)
		}

		rt, err := runtimes().New(context.Background(), name)
		err = common.AddNotFoundErrIfCond(err, rt == nil)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "runtime"); err == nil {
			common.Render(rt.Get())
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
