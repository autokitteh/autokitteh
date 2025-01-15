package orgs

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var listMembersCmd = common.StandardCommand(&cobra.Command{
	Use:     "list-members <org id or name>",
	Short:   "List members in an org",
	Aliases: []string{"lsm"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		r := resolver.Resolver{Client: common.Client()}

		oid, err := r.Org(ctx, args[0])
		if err != nil {
			return fmt.Errorf("org: %w", err)
		}

		members, err := orgs().ListMembers(ctx, oid)

		err = common.AddNotFoundErrIfCond(err, len(members) > 0)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "members"); err == nil {
			common.RenderList(members)
		}

		return err
	},
})
