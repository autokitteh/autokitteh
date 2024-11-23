package projects

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var name, owner string

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create [--name project-name] [--owner owner-email-or-id]",
	Short:   "Create new project",
	Aliases: []string{"c"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		nameSym, err := sdktypes.ParseSymbol(name)
		if err != nil {
			return fmt.Errorf("parse project name: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		r := resolver.Resolver{Client: common.Client()}
		_, oid, err := r.Owner(ctx, owner)

		p := sdktypes.NewProject().WithName(nameSym).WithOwnerID(oid)

		id, err := projects().Create(ctx, p)
		if err != nil {
			return fmt.Errorf("create project: %w", err)
		}

		common.RenderKV("project_id", id)
		return nil
	},
})

func init() {
	createCmd.Flags().StringVarP(&name, "name", "n", "", "project name")
	createCmd.Flags().StringVarP(&owner, "owner", "o", "", "project owner email or id")
}
