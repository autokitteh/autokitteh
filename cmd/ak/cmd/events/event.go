package events

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	valuesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Flags shared by the "create", "dispatch", and "list" subcommands.
var (
	filename, integration, connection, trigger, eventType string

	data, memos []string
)

var eventCmd = common.StandardCommand(&cobra.Command{
	Use:     "event",
	Short:   "Events: save, get, list, (re)dispatch, record",
	Aliases: []string{"evt"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(eventCmd)
}

func init() {
	// Subcommands.
	eventCmd.AddCommand(dispatchCmd)
	eventCmd.AddCommand(getCmd)
	eventCmd.AddCommand(listCmd)
	eventCmd.AddCommand(redispatchCmd)
	eventCmd.AddCommand(saveCmd)
	eventCmd.AddCommand(verifyCmd)
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
