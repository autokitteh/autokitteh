package events

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/events/records"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	valuesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Flags shared by the "create", "dispatch", and "list" subcommands.
var (
	filename, integration, connectionToken, eventType, originalEventID string

	data, memos []string
)

var eventsCmd = common.StandardCommand(&cobra.Command{
	Use:     "events",
	Short:   "Event management commands",
	Aliases: []string{"event", "evt", "ev"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(eventsCmd)
}

func init() {
	// Subcommands.
	eventsCmd.AddCommand(saveCmd)
	eventsCmd.AddCommand(dispatchCmd)
	eventsCmd.AddCommand(redispatchCmd)
	eventsCmd.AddCommand(getCmd)
	eventsCmd.AddCommand(listCmd)

	records.AddSubcommands(eventsCmd)
}

func events() sdkservices.Events {
	return common.Client().Events()
}

func parseDataKeyValue(kv string) (string, *valuesv1.Value, error) {
	k, v, ok := strings.Cut(kv, "=")
	if !ok {
		return "", nil, fmt.Errorf(`invalid data pair %q, expected "key=value"`, kv)
	}

	vw := sdktypes.ValueWrapper{}
	value, err := vw.Wrap(v)
	if err != nil {
		return "", nil, fmt.Errorf("invalid data value %q: %w", v, err)
	}
	return k, value.ToProto(), nil
}

func parseMemoKeyValue(kv string) (string, string, error) {
	k, v, ok := strings.Cut(kv, "=")
	if !ok {
		return "", "", fmt.Errorf(`invalid memo pair %q, expected "key=value"`, kv)
	}

	return k, v, nil
}
