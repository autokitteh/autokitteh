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
		f := sdkservices.ListEventsFilter{
			IntegrationToken: connectionToken,
			OriginalID:       originalEventID,
			EventType:        eventType,
		}

		if integration != "" {
			r := resolver.Resolver{Client: common.Client()}
			i, iid, err := r.IntegrationNameOrID(integration)
			if err != nil {
				return err
			}
			if i == nil {
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

		if len(es) == 0 {
			return common.FailNotFound(cmd, "events")
		}

		common.RenderList(es)
		return nil
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&integration, "integration", "i", "", "integration name or ID")
	listCmd.Flags().StringVarP(&connectionToken, "connection-token", "t", "", "connection token")
	listCmd.Flags().StringVarP(&eventType, "event-type", "e", "", "event type")
	listCmd.Flags().StringVarP(&originalEventID, "original-event-id", "o", "", "original event ID")

	listCmd.MarkFlagsOneRequired("integration", "connection-token", "event-type", "original-event-id")

	common.AddFailIfNotFoundFlag(listCmd)
}
