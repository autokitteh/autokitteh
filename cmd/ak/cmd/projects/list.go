package projects

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--org org-name-or-id] [--fail]",
	Short:   "List all projects",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		r := resolver.Resolver{Client: common.Client()}

		oid, err := r.Org(ctx, org)
		if err != nil {
			return common.WrapError(err, "org")
		}

		ps, err := projects().List(ctx, oid)
		err = common.AddNotFoundErrIfCond(err, len(ps) > 0)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "projects"); err == nil {
			common.RenderList(ps)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)

	listCmd.Flags().StringVarP(&org, "org", "o", "", "project org name or id")
}
