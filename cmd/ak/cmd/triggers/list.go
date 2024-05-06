package triggers

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--env=...] [--conection=...] [--fail]",
	Short:   "List all event triggers",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		f := sdkservices.ListTriggersFilter{}

		if env != "" {
			_, eid, err := r.EnvNameOrID(env, "")
			if err != nil {
				return err
			}
			f.EnvID = eid
		}

		c, cid, err := r.ConnectionNameOrID(connection, "")
		if err != nil {
			if errors.As(err, resolver.NotFoundErrorType) {
				err = common.NewExitCodeError(common.NotFoundExitCode, err)
			}
			return err
		}
		if cid.IsValid() && !c.IsValid() {
			err = fmt.Errorf("connection ID %q not found", connection)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}
		f.ConnectionID = cid

		ctx, cancel := common.LimitedContext()
		defer cancel()

		ts, err := triggers().List(ctx, f)
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "triggers", len(ts) > 0); err != nil {
			return err
		}

		common.RenderList(ts)
		return nil
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	listCmd.Flags().StringVarP(&connection, "connection", "n", "", "connection name or ID")

	common.AddFailIfNotFoundFlag(listCmd)
}
