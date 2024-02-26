package runtimes

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <runtime name>",
	Short:   "Get runtime engine details",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"g"},

	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := sdktypes.ParseName(args[0])
		if err != nil {
			return fmt.Errorf("name: %w", err)
		}

		rt, err := runtimes().New(context.Background(), name)
		if err != nil {
			return err
		}

		// QUESTION: Why not "common.FailIfNotFound" like other "get" commands?
		// If we do add it, don't forget to add "[--fail]" to "Use" above,
		// and an "init" function calling "common.AddFailIfNotFoundFlag".
		if rt == nil {
			return nil
		}

		common.Render(rt.Get())
		return nil
	},
})
