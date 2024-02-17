package projects

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var buildCmd = common.StandardCommand(&cobra.Command{
	Use:     "build <project name or ID>",
	Short:   `Build project (see also the "build" subcommands)`,
	Aliases: []string{"b"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		p, pid, err := r.ProjectNameOrID(args[0])
		if err != nil {
			return err
		}
		if p == nil {
			err = fmt.Errorf("project %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		bid, err := projects().Build(ctx, pid)
		if err != nil {
			return fmt.Errorf("build project: %w", err)
		}

		common.RenderKVIfV("build_id", bid)
		return nil
	},
})
