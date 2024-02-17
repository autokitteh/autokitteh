package vars

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var secret bool

var setCmd = common.StandardCommand(&cobra.Command{
	Use:     "set <key> <value> [--secret] <--env=...> [--project=...]",
	Short:   "Set environment variable",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(2),

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

		ev, err := sdktypes.EnvVarFromProto(&sdktypes.EnvVarPB{
			EnvId:    id.String(),
			Name:     args[0],
			Value:    args[1],
			IsSecret: secret,
		})
		if err != nil {
			return fmt.Errorf("invalid environment variable: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		if err := envs().SetVar(ctx, ev); err != nil {
			return fmt.Errorf("set environment variable: %w", err)
		}
		return nil
	},
})

func init() {
	// Command-specific flags.
	setCmd.Flags().BoolVarP(&secret, "secret", "s", false, "this is a secret")
}
