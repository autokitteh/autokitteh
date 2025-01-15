package orgs

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getMemberCmd = common.StandardCommand(&cobra.Command{
	Use:     "get-member <org id or name> <user id>",
	Short:   "Get org member",
	Aliases: []string{"gm"},
	Args:    cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		r := resolver.Resolver{Client: common.Client()}

		oid, err := r.Org(ctx, args[0])
		if err != nil {
			return fmt.Errorf("org: %w", err)
		}

		uid, err := r.UserID(ctx, args[1])
		if err != nil {
			return fmt.Errorf("user: %w", err)
		}

		s, err := orgs().GetMemberStatus(ctx, oid, uid)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "member"); err == nil {
			common.RenderKVIfV("member", s)
		}
		return err
	},
})

func init() {
	common.AddFailIfNotFoundFlag(getMemberCmd)
}
