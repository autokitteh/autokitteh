package projects

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var name, org string

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create [--name project-name] [--org org-name-or-id]",
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
		oid, err := r.Org(ctx, org)
		if err != nil {
			return fmt.Errorf("resolve org: %w", err)
		}

		p := sdktypes.NewProject().WithName(nameSym).WithOrgID(oid)

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
	createCmd.Flags().StringVarP(&org, "org", "o", "", "project org name or id")
}
