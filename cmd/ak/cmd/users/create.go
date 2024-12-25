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
	disabled           bool
)

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create --email email [--display-name display-name] [--disabled]",
	Short:   "Create new user",
	Aliases: []string{"c"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		u := sdktypes.NewUser(email).WithDisplayName(displayName).WithDisabled(disabled)

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

	createCmd.Flags().StringVarP(&email, "display-name", "t", "", "user's display name")
	createCmd.Flags().BoolVarP(&disabled, "disabled", "d", false, "is user disabled")
}
