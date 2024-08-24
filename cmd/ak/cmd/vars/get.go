package vars

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var reveal bool

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get [k1 [k2 ...]] <--env=... | --connection=...> [--project=...] [--reveal]",
	Short:   "Get variable(s)",
	Aliases: []string{"g"},
	Args:    cobra.ArbitraryArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveScopeID()
		if err != nil {
			return err
		}

		ks, err := kittehs.TransformError(args, sdktypes.StrictParseSymbol)
		if err != nil {
			return fmt.Errorf("invalid value name: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		vs, err := vars().Get(ctx, id, ks...)
		if err != nil {
			return fmt.Errorf("get variable(s): %w", err)
		}

		common.RenderList(kittehs.Transform(vs, func(v sdktypes.Var) V { return V{v} }))
		return nil
	},
})

type V struct{ sdktypes.Var }

func (v V) ToString() string {
	vv := "<secret>"

	if !v.Var.IsSecret() || reveal {
		vv = strconv.Quote(v.Var.Value())
	}

	return fmt.Sprintf("%v=%s", v.Var.Name(), vv)
}

var _ common.ToStringer = V{}

func init() {
	getCmd.Flags().BoolVar(&reveal, "reveal", false, "reveal secret values")
}
