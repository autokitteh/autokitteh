package deployments

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var activateCmd = common.StandardCommand(&cobra.Command{
	Use:     "activate <deployment ID>",
	Short:   "Activate deployment",
	Aliases: []string{"a"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		d, id, err := r.DeploymentID(args[0])
		if err != nil {
			return err
		}
		if d == nil {
			err = fmt.Errorf("deployment %q not found", args[0])
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		if err := deployments().Activate(ctx, id); err != nil {
			return fmt.Errorf("activate deployment: %w", err)
		}

		return nil
	},
})
