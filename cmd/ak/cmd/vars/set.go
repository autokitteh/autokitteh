package vars

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var secret bool

var setCmd = common.StandardCommand(&cobra.Command{
	Use:     "set <key> <value> [--secret] <--env=.. | --connection=....> [--project=...]",
	Short:   "Set variable",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveScopeID()
		if err != nil {
			return err
		}

		n, err := sdktypes.StrictParseSymbol(args[0])
		if err != nil {
			return fmt.Errorf("invalid variable name %q: %w", args[0], err)
		}

		ev := sdktypes.NewVar(n, args[1], secret).WithScopeID(id)
		if err != nil {
			return fmt.Errorf("invalid variable: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		if err := vars().Set(ctx, ev); err != nil {
			return fmt.Errorf("set variable: %w", err)
		}
		return nil
	},
})

func init() {
	// Command-specific flags.
	setCmd.Flags().BoolVarP(&secret, "secret", "s", false, "this is a secret")
}
