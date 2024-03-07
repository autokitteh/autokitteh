package projects

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var name string

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create [--name project-name]",
	Short:   "Create new project",
	Aliases: []string{"c"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := sdktypes.ParseSymbol(name); err != nil {
			return fmt.Errorf("invalid project name: %w", err)
		}

		p, err := sdktypes.ProjectFromProto(&sdktypes.ProjectPB{Name: name})
		if err != nil {
			return err
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		id, err := projects().Create(ctx, p)
		if err != nil {
			return fmt.Errorf("create project: %w", err)
		}

		if name != "" {
			common.RenderKV("project_id", id)
			return nil
		}

		if p, err = projects().GetByID(ctx, id); err != nil {
			return fmt.Errorf("get project: %w", err)
		}

		common.RenderKV("project", p)
		return nil
	},
})

func init() {
	createCmd.Flags().StringVarP(&name, "name", "n", "", "project name")
}
