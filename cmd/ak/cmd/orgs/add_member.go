package orgs

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	active bool
	roles  []string
)

var addMemberCmd = common.StandardCommand(&cobra.Command{
	Use:     "add-member <org id or name> <user id> [--active]",
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

		rs, err := kittehs.TransformError(roles, sdktypes.ParseSymbol)
		if err != nil {
			return fmt.Errorf("roles: %w", err)
		}

		m := sdktypes.NewOrgMember(oid, uid).WithStatus(s).WithRoles(rs...)

		if err := orgs().AddMember(ctx, m); err != nil {
			return fmt.Errorf("add member: %w", err)
		}

		return nil
	},
})

func init() {
	addMemberCmd.Flags().BoolVarP(&active, "active", "a", false, "add member as active")
	addMemberCmd.Flags().StringSliceVarP(&roles, "role", "r", nil, "add member with role")
}
