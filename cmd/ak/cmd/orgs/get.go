package orgs

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:   "get <id> [--fail]",
	Short: "Get org details",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		oid, err := r.Org(ctx, args[0])
		if err != nil {
			return err
		}

		o, err := orgs().GetByID(ctx, oid)
		err = common.AddNotFoundErrIfCond(err, oid.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "user"); err == nil {
			common.RenderKVIfV("org", o)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
