package users

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var status string

var updateCmd = common.StandardCommand(&cobra.Command{
	Use:   "update [email or id] [--display-name display-name] [--disabled] [--status status]",
	Short: "Update a user",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		r := resolver.Resolver{Client: common.Client()}

		var uid sdktypes.UserID

		if len(args) > 0 {
			var err error
			if _, uid, err = r.User(ctx, args[0]); err != nil {
				return fmt.Errorf("resolve user: %w", err)
			}
		}

		u := sdktypes.NewUser().WithID(uid)

		fm := &sdktypes.FieldMask{}

		if cmd.Flags().Changed("display-name") {
			fm.Paths = append(fm.Paths, "display_name")
			u = u.WithDisplayName(displayName)
		}

		if cmd.Flags().Changed("status") {
			fm.Paths = append(fm.Paths, "status")

			s, err := sdktypes.ParseUserStatus(status)
			if err != nil {
				return fmt.Errorf("parse status: %w", err)
			}

			u = u.WithStatus(s)
		}

		if err := users().Update(ctx, u, fm); err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		return nil
	},
})

func init() {
	updateCmd.Flags().StringVarP(&displayName, "display-name", "t", "", "user's display name")
	updateCmd.Flags().StringVarP(&status, "status", "", "", "user status")
}
