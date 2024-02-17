package mappings

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--env=...] [--conection=...] [--fail]",
	Short:   "List all connection mappings",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		f := sdkservices.ListMappingsFilter{}

		if env != "" {
			_, eid, err := r.EnvNameOrID(env, "")
			if err != nil {
				return err
			}
			f.EnvID = eid
		}

		c, cid, err := r.ConnectionNameOrID(connection)
		if err != nil {
			if errors.As(err, resolver.NotFoundErrorType) {
				err = common.NewExitCodeError(common.NotFoundExitCode, err)
			}
			return err
		}
		if cid != nil && c == nil {
			err = fmt.Errorf("connection ID %q not found", connection)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}
		f.ConnectionID = cid

		ctx, cancel := common.LimitedContext()
		defer cancel()

		ms, err := mappings().List(ctx, f)
		if err != nil {
			return err
		}

		if len(ms) == 0 {
			var dummy *sdktypes.Mapping
			return common.FailIfNotFound(cmd, "mappings", dummy)
		}

		common.RenderList(ms)
		return nil
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	listCmd.Flags().StringVarP(&connection, "connection", "n", "", "connection name or ID")

	common.AddFailIfNotFoundFlag(listCmd)
}
