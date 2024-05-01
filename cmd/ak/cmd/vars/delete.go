package vars

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var deleteCmd = common.StandardCommand(&cobra.Command{
	Use:     "delete <key> <--env=... | --connection=...> [--project=...]",
	Short:   "Delete environment variable",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveScopeID()
		if err != nil {
			return err
		}

		k, err := sdktypes.StrictParseSymbol(args[0])
		if err != nil {
			return fmt.Errorf("invalid variable name %q: %w", args[0], err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		if err := vars().Delete(ctx, id, k); err != nil {
			return fmt.Errorf("remove environment variable: %w", err)
		}

		return nil
	},
})
