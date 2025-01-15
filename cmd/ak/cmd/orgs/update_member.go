package orgs

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var status string

var updateMemberCmd = common.StandardCommand(&cobra.Command{
	Use:     "update-member <org id or name> <user id> [--status status]",
	Short:   "Update org member",
	Aliases: []string{"um"},
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

		s, err := sdktypes.ParseOrgMemberStatus(status)
		if err != nil {
			return fmt.Errorf("status: %w", err)
		}

		if err := orgs().UpdateMemberStatus(ctx, oid, uid, s); err != nil {
			return fmt.Errorf("update member: %w", err)
		}

		return nil
	},
})

func init() {
	updateMemberCmd.Flags().StringVarP(&status, "status", "s", "", "new member status")
	kittehs.Must0(updateMemberCmd.MarkFlagRequired("status"))
}
