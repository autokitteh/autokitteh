package events

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var saveCmd = common.StandardCommand(&cobra.Command{
	Use:     "save [--from-file=...] [override flags]",
	Short:   "Save new event",
	Aliases: []string{"s"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		var event sdktypes.Event
		pb := &sdktypes.EventPB{}

		if filename != "" {
			text, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
			if err := json.Unmarshal(text, &event); err != nil {
				return fmt.Errorf("unmarshal JSON in %q: %w", filename, err)
			}
			pb = event.ToProto()
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
			pb.IntegrationId = iid.String()
		}
		if connectionToken != "" {
			pb.IntegrationToken = connectionToken
		}
		if len(data) > 0 {
			m, err := kittehs.ListToMapError(data, parseDataKeyValue)
			if err != nil {
				return err
			}
			pb.Data = m
		}
		if len(memos) > 0 {
			memoMap, err := kittehs.ListToMapError(memos, parseMemoKeyValue)
			if err != nil {
				return err
			}
			pb.Memo = memoMap
		}

		e, err := sdktypes.EventFromProto(pb)
		if err != nil {
			return fmt.Errorf("invalid event: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		eid, err := events().Save(ctx, e)
		if err != nil {
			return fmt.Errorf("save event: %w", err)
		}

		common.RenderKV("event_id", eid)
		return nil
	},
})

func init() {
	// Command-specific flags.
	saveCmd.Flags().StringVarP(&filename, "from-file", "f", "", "load event data from file")
	kittehs.Must0(saveCmd.MarkFlagFilename("from-file"))

	saveCmd.Flags().StringVarP(&integration, "integration", "i", "", "integration name or ID")
	saveCmd.Flags().StringVarP(&connectionToken, "connection-token", "t", "", "connection token")
	saveCmd.Flags().StringVarP(&eventType, "event-type", "e", "", "event type")
	saveCmd.Flags().StringSliceVarP(&data, "data", "d", nil, `zero or more "key=value" pairs`)
	saveCmd.Flags().StringSliceVarP(&memos, "memo", "m", nil, `zero or more "key=value" pairs`)
}
