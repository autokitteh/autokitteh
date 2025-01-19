package users

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	email, displayName string
	active             bool
)

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create --email email [--display-name display-name] [--active]",
	Short:   "Create new user",
	Aliases: []string{"c"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		s := sdktypes.UserStatusInvited
		if active {
			s = sdktypes.UserStatusActive
		}

		u := sdktypes.NewUser().WithEmail(email).WithDisplayName(displayName).WithStatus(s)

		id, err := users().Create(ctx, u)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		common.RenderKV("user_id", id)
		return nil
	},
})

func init() {
	createCmd.Flags().StringVarP(&email, "email", "e", "", "user's email")
	kittehs.Must0(createCmd.MarkFlagRequired("email"))

	createCmd.Flags().StringVarP(&displayName, "display-name", "t", "", "user's display name")
	createCmd.Flags().BoolVarP(&active, "active", "a", false, "is user active")
}
