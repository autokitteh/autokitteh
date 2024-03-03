package envs

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create <env name> <--project=...>",
	Short:   "Create new execution environment",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"c"},

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		p, pid, err := r.ProjectNameOrID(project)
		if err != nil {
			return err
		}
		if !p.IsValid() {
			err = fmt.Errorf("project %q not found", project)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		e, err := sdktypes.EnvFromProto(&sdktypes.EnvPB{ProjectId: pid.String(), Name: args[0]})
		if err != nil {
			return fmt.Errorf("invalid environment: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		eid, err := envs().Create(ctx, e)
		if err != nil {
			return fmt.Errorf("create environment: %w", err)
		}

		common.RenderKV("env_id", eid)
		return nil
	},
})
