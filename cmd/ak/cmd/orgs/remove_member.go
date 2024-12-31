package orgs

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var removeMemberCmd = common.StandardCommand(&cobra.Command{
	Use:     "remove-member <org id> <user id>",
	Short:   "Remove org member",
	Aliases: []string{"rmm"},
	Args:    cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		r := resolver.Resolver{Client: common.Client()}

		oid, err := r.Org(ctx, args[0])
		if err != nil {
			return fmt.Errorf("org: %w", err)
		}

		_, uid, err := r.User(ctx, args[1])
		if err != nil {
			return fmt.Errorf("user: %w", err)
		}

		if err := orgs().RemoveMember(ctx, oid, uid); err != nil {
			return fmt.Errorf("add member: %w", err)
		}

		return nil
	},
})
