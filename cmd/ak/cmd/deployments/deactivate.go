package deployments

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var deactivateCmd = common.StandardCommand(&cobra.Command{
	Use:     "deactivate <deployment ID>",
	Short:   "Deactivate deployment",
	Aliases: []string{"de"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		d, id, err := r.DeploymentID(ctx, args[0])
		err = common.AddNotFoundErrIfCond(err, d.IsValid())
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "deployment"); err != nil {
			return err
		}

		if err := deployments().Deactivate(ctx, id); err != nil {
			return fmt.Errorf("deactivate deployment: %w", err)
		}

		return nil
	},
})
