package vars

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var removeCmd = common.StandardCommand(&cobra.Command{
	Use:     "remove <key> <--env=...> [--project=...]",
	Short:   "Remove environment variable",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		e, id, err := r.EnvNameOrID(env, project)
		if err != nil {
			return err
		}
		if !e.IsValid() {
			err = fmt.Errorf("environment %q not found", env)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		k, err := sdktypes.StrictParseSymbol(args[0])
		if err != nil {
			return fmt.Errorf("invalid variable name %q: %w", args[0], err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		if err := envs().RemoveVar(ctx, id, k); err != nil {
			return fmt.Errorf("remove environment variable: %w", err)
		}

		return nil
	},
})
