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
		ctx, cancel := common.LimitedContext()
		defer cancel()

		d, id, err := r.DeploymentID(ctx, args[0])
		err = common.AddNotFoundErrIfCond(err, d.IsValid())
		if err = common.FailIfError2(cmd, err, "deployment"); err != nil {
			return err
		}

		if err := deployments().Activate(ctx, id); err != nil {
			return fmt.Errorf("activate deployment: %w", err)
		}

		return nil
	},
})
