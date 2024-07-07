package deployments

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var drainCmd = common.StandardCommand(&cobra.Command{
	Use:     "drain <deployment ID>",
	Short:   "Drain deployment",
	Aliases: []string{"dr"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		d, id, err := r.DeploymentID(ctx, args[0])
		if err != nil {
			return err
		}
		if !d.IsValid() {
			err = fmt.Errorf("deployment %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		if err := deployments().Drain(ctx, id); err != nil {
			return fmt.Errorf("drain deployment: %w", err)
		}

		return nil
	},
})
