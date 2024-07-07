package envs

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--project=...] [--fail]",
	Short:   "List all execution environments",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			p   sdktypes.Project
			pid sdktypes.ProjectID
			err error
		)

		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		if project != "" {
			p, pid, err = r.ProjectNameOrID(ctx, project)
			if err != nil {
				return err
			}
			if !p.IsValid() {
				err = fmt.Errorf("project %q not found", project)
				return common.NewExitCodeError(common.NotFoundExitCode, err)
			}
		}

		es, err := envs().List(ctx, pid)
		if err != nil {
			return fmt.Errorf("list environments: %w", err)
		}

		if err := common.FailIfNotFound(cmd, "environments", len(es) > 0); err != nil {
			return err
		}

		common.RenderList(es)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)
}
