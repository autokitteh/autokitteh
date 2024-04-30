package sessions

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Default flag value shared by the "start", "restart", and "watch" subcommands.
const (
	defaultPollInterval = 1 * time.Second
)

var (
	// Flags shared by the "list" and "start" subcommands.
	deploymentID, env, eventID string

	// Flags shared by the "start", "restart", and "watch" subcommands.
	pollInterval time.Duration
	watchTimeout time.Duration
	watch, quiet bool
	noTimestamps bool
)

var sessionCmd = common.StandardCommand(&cobra.Command{
	Use:     "session",
	Short:   "Runtime sessions: (re)start, get, list, log, watch, test, stop, delete",
	Aliases: []string{"ses"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(sessionCmd)
}

func init() {
	// Subcommands.
	sessionCmd.AddCommand(deleteCmd)
	sessionCmd.AddCommand(getCmd)
	sessionCmd.AddCommand(listCmd)
	sessionCmd.AddCommand(logCmd)
	sessionCmd.AddCommand(restartCmd)
	sessionCmd.AddCommand(startCmd)
	sessionCmd.AddCommand(stopCmd)
	sessionCmd.AddCommand(testCmd)
	sessionCmd.AddCommand(watchCmd)
}

func sessions() sdkservices.Sessions {
	return common.Client().Sessions()
}

func latestSessionID() (string, error) {
	ctx, cancel := common.LimitedContext()
	defer cancel()

	result, err := sessions().List(ctx, sdkservices.ListSessionsFilter{})
	if err != nil {
		return "", fmt.Errorf("list sessions: %w", err)
	}

	if len(result.Sessions) == 0 {
		return "", common.NewExitCodeError(common.NotFoundExitCode, errors.New("sessions not found"))
	}

	latest := result.Sessions[0]
	for _, s := range result.Sessions[1:] {
		if s.ToProto().CreatedAt.AsTime().After(latest.ToProto().CreatedAt.AsTime()) {
			latest = s
		}
	}
	return string(latest.ToProto().SessionId), nil
}
