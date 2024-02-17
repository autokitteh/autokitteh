package vars

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var revealCmd = common.StandardCommand(&cobra.Command{
	Use:     "reveal <key> <--env=...> [--project=...]",
	Short:   "Reveal secret environment variable",
	Aliases: []string{"r"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		e, id, err := r.EnvNameOrID(env, project)
		if err != nil {
			return err
		}
		if e == nil {
			err = fmt.Errorf("environment %q not found", env)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		k, err := sdktypes.StrictParseSymbol(args[0])
		if err != nil {
			return fmt.Errorf("invalid value name %q: %w", args[0], err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		v, err := envs().RevealVar(ctx, id, k)
		if err != nil {
			return fmt.Errorf("reveal environment variable: %w", err)
		}

		common.Render(v)
		return nil
	},
})
