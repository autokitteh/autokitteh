package projects

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create <project name>",
	Short:   "Create new project",
	Aliases: []string{"c"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := sdktypes.ProjectFromProto(&sdktypes.ProjectPB{Name: args[0]})
		if err != nil {
			return err
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		id, err := projects().Create(ctx, p)
		if err != nil {
			return fmt.Errorf("create project: %w", err)
		}

		common.RenderKV("project_id", id)
		return nil
	},
})
