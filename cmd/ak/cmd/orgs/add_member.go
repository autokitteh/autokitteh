package orgs

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var active bool

var addMemberCmd = common.StandardCommand(&cobra.Command{
	Use:     "add-member <org id> <user id> [--active]",
	Short:   "Add org member",
	Aliases: []string{"am"},
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

		s := sdktypes.OrgMemberStatusInvited
		if active {
			s = sdktypes.OrgMemberStatusActive
		}

		if err := orgs().AddMember(ctx, oid, uid, s); err != nil {
			return fmt.Errorf("add member: %w", err)
		}

		return nil
	},
})

func init() {
	addMemberCmd.Flags().BoolVarP(&active, "active", "a", false, "add member as active")
}
