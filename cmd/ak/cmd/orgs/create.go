package orgs

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var displayName string

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create [--display-name display-name]",
	Short:   "Create new org",
	Aliases: []string{"c"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		o := sdktypes.NewOrg().WithDisplayName(displayName)

		id, err := orgs().Create(ctx, o)
		if err != nil {
			return fmt.Errorf("create org: %w", err)
		}

		common.RenderKV("org_id", id)
		return nil
	},
})

func init() {
	createCmd.Flags().StringVarP(&displayName, "display-name", "t", "", "org's display name")
}
