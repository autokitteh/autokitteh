package vars

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get [k1 [k2 ...]] <--env=...> [--project=...]",
	Short:   "Get environment variable(s)",
	Aliases: []string{"g"},
	Args:    cobra.ArbitraryArgs,

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

		ks, err := kittehs.TransformError(args, sdktypes.StrictParseSymbol)
		if err != nil {
			return fmt.Errorf("invalid value name: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		vs, err := envs().GetVars(ctx, ks, id)
		if err != nil {
			return fmt.Errorf("get environment variable(s): %w", err)
		}

		common.RenderList(kittehs.Transform(vs, func(v sdktypes.EnvVar) V { return V{v} }))
		return nil
	},
})

type V struct{ sdktypes.EnvVar }

func (v V) Text() string {
	vv := "<secret>"

	if !sdktypes.IsEnvVarSecret(v.EnvVar) {
		vv = strconv.Quote(sdktypes.GetEnvVarValue(v.EnvVar))
	}

	return fmt.Sprintf("%s=%s", sdktypes.GetEnvVarName(v.EnvVar), vv)
}

var _ common.Texter = V{}
