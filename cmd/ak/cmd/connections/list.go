package connections

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--integration=...] [--connection-token=...] [--fail]",
	Short:   "List all connections",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		f := sdkservices.ListConnectionsFilter{}

		if integration != "" {
			r := resolver.Resolver{Client: common.Client()}
			_, iid, err := r.IntegrationNameOrID(integration)
			if err != nil {
				return err
			}
			f.IntegrationID = iid
		}

		f.IntegrationToken = connectionToken

		ctx, cancel := common.LimitedContext()
		defer cancel()

		cs, err := connections().List(ctx, f)
		if err != nil {
			return fmt.Errorf("list connections: %w", err)
		}

		if len(cs) == 0 {
			var dummy *sdktypes.Connection
			return common.FailIfNotFound(cmd, "connections", dummy)
		}

		common.RenderList(cs)
		return nil
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&integration, "integration", "i", "", "integration name or ID")
	listCmd.Flags().StringVarP(&connectionToken, "connection-token", "t", "", "connection token")

	common.AddFailIfNotFoundFlag(listCmd)
}
