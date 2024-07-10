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

		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

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

		if connection != "" {
			_, cid, err := r.ConnectionNameOrID(ctx, args[0], "")
			if err != nil {
				return err
			}
			if !cid.IsValid() {
				return fmt.Errorf("connection %q not found", connection)
			}

			pb.ConnectionId = cid.String()
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

	saveCmd.Flags().StringVarP(&connection, "connection", "c", "", "connection name or ID")
	saveCmd.Flags().StringVarP(&eventType, "event-type", "e", "", "event type")
	saveCmd.Flags().StringSliceVarP(&data, "data", "d", nil, `zero or more "key=value" pairs`)
	saveCmd.Flags().StringSliceVarP(&memos, "memo", "m", nil, `zero or more "key=value" pairs`)
}
