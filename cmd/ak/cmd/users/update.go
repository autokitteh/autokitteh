package users

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var enabled bool

var updateCmd = common.StandardCommand(&cobra.Command{
	Use:     "update [email or id] [--email email] [--display-name display-name] [--disabled] [--enabled]",
	Short:   "Update a user",
	Aliases: []string{"u"},
	Args:    cobra.ExactArgs(1),

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

		if cmd.Flags().Changed("email") {
			fm.Paths = append(fm.Paths, "email")
			u = u.WithEmail(email)
		}

		if cmd.Flags().Changed("display-name") {
			fm.Paths = append(fm.Paths, "display_name")
			u = u.WithDisplayName(displayName)
		}

		if cmd.Flags().Changed("disabled") {
			fm.Paths = append(fm.Paths, "disabled")
			u = u.WithDisabled(true)
		}

		if cmd.Flags().Changed("enabled") {
			fm.Paths = append(fm.Paths, "disabled")
			u = u.WithDisabled(false)
		}

		if err := users().Update(ctx, u, fm); err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		return nil
	},
})

func init() {
	updateCmd.Flags().StringVarP(&email, "email", "e", "", "user's email")
	updateCmd.Flags().StringVarP(&displayName, "display-name", "t", "", "user's display name")
	updateCmd.Flags().BoolVarP(&disabled, "disabled", "d", false, "is user disabled")
	updateCmd.Flags().BoolVarP(&enabled, "enabled", "", false, "is user enabled")
	updateCmd.MarkFlagsMutuallyExclusive("disabled", "enabled")
}
