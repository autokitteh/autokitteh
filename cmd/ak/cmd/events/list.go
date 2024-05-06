package events

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [filter flags] [--fail]",
	Short:   "List all events",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		var f sdkservices.ListEventsFilter

		if connection != "" {
			r := resolver.Resolver{Client: common.Client()}
			_, cid, err := r.ConnectionNameOrID(args[0], "")
			if err != nil {
				return err
			}
			if !cid.IsValid() {
				return fmt.Errorf("connection %q not found", connection)
			}
			f.ConnectionID = cid
		}

		if integration != "" {
			r := resolver.Resolver{Client: common.Client()}
			i, iid, err := r.IntegrationNameOrID(integration)
			if err != nil {
				return err
			}
			if !i.IsValid() {
				return fmt.Errorf("integration %q not found", integration)
			}
			f.IntegrationID = iid
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		es, err := events().List(ctx, f)
		if err != nil {
			return fmt.Errorf("list events: %w", err)
		}

		if err := common.FailIfNotFound(cmd, "events", len(es) > 0); err != nil {
			return err
		}

		common.RenderList(es)
		return nil
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&integration, "integration", "i", "", "integration name or ID")
	listCmd.Flags().StringVarP(&connection, "connection", "c", "", "connection name or ID")
	listCmd.Flags().StringVarP(&eventType, "event-type", "e", "", "event type")

	listCmd.MarkFlagsOneRequired("integration", "connection")

	common.AddFailIfNotFoundFlag(listCmd)
}
