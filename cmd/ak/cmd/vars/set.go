package vars

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var secret, optional bool

var setCmd = common.StandardCommand(&cobra.Command{
	Use:     "set <key> [<value>] [--secret] [--optional] <--env=.. | --connection=....> [--project=...]",
	Short:   "Set variable",
	Long:    `Set a variable. If <value> is "-", it is read from stdin.`,
	Aliases: []string{"s"},
	Args:    cobra.RangeArgs(1, 2),

	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveScopeID()
		if err != nil {
			return err
		}

		n, err := sdktypes.StrictParseSymbol(args[0])
		if err != nil {
			return fmt.Errorf("invalid variable name %q: %w", args[0], err)
		}

		var value string

		isValueSet := len(args) > 1

		if isValueSet {
			if args[1] != "-" {
				value = args[1]
			} else {
				const maxVarSize = 1 << 20 // 1MB
				r := io.LimitReader(os.Stdin, maxVarSize)
				data, err := io.ReadAll(r)
				if err != nil {
					return err
				}
				value = string(data)
			}
		} else {
			// get current var in order not to overwrite it with empty.

			ctx, cancel := common.LimitedContext()
			defer cancel()

			vs, err := vars().Get(ctx, id, n)
			if err != nil {
				return fmt.Errorf("get variable: %w", err)
			}

			if len(vs) > 0 {
				value = vs[0].Value()
			}
		}

		ev := sdktypes.NewVar(n).SetValue(value).SetSecret(secret).SetOptional(optional).WithScopeID(id)
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
	setCmd.Flags().BoolVarP(&secret, "secret", "s", false, "this value is secret")
	setCmd.Flags().BoolVarP(&optional, "optional", "o", false, "this value is optional")
}
